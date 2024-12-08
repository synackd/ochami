// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/OpenCHAMI/ochami/internal/client"
	"github.com/OpenCHAMI/ochami/internal/config"
	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/internal/version"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	progName             = "ochami"
	defaultLogFormat     = "json"
	defaultLogLevel      = "warning"
	defaultConfigFormat  = "yaml"
	defaultPayloadFormat = "json"
)

var (
	configFile   string
	configFormat string
	logLevel     string
	logFormat    string

	// These are only used by 'bss' and 'smd' subcommands.
	baseURI    string
	cacertPath string
	token      string
	insecure   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     progName,
	Args:    cobra.NoArgs,
	Short:   "Command line interface for interacting with OpenCHAMI services",
	Long:    "",
	Version: version.Version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to execute root command")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(
		InitConfig,
		InitLogging,
		InitialStatus,
	)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to configuration file to use")
	rootCmd.PersistentFlags().StringVar(&configFormat, "config-format", defaultConfigFormat, "format of configuration file; if none passed, tries to infer from file extension")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", defaultLogFormat, "log format (json,rfc3339,basic)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", defaultLogLevel, "set verbosity of logs (info,warning,debug)")
	rootCmd.PersistentFlags().String("cluster", "", "name of cluster whose config to use for this command")
	rootCmd.PersistentFlags().StringVarP(&baseURI, "base-uri", "u", "", "base URI for OpenCHAMI services")
	rootCmd.PersistentFlags().StringVar(&cacertPath, "cacert", "", "path to root CA certificate in PEM format")
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "access token to present for authentication")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "do not verify TLS certificates")
	rootCmd.PersistentFlags().Bool("ignore-config", false, "do not use any config file")

	// Either use cluster from config file or specify details on CLI
	bssCmd.MarkFlagsMutuallyExclusive("cluster", "base-uri")

	checkBindError(viper.BindPFlag("log.format", rootCmd.PersistentFlags().Lookup("log-format")))
	checkBindError(viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level")))
}

func checkBindError(e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to bind key to flag: %v\n", progName, e)
	}
}

func InitLogging() {
	// Set log level verbosity based on config file (log.level) or how many --log-level.
	// The command line option overrides the config file option.
	logCfg := viper.Sub("log")
	if logCfg == nil {
		fmt.Fprintf(os.Stderr, "%s: failed to read logging config", progName)
		os.Exit(1)
	}

	// Viper's BindPFlag does not currently work with binding to subkeys.
	// (See: https://github.com/spf13/viper/issues/368)
	// Therefore, we must manually check if the flag was set. If not, check if
	// config file option was set. If not, use default value.
	//
	// These if statements should be removed when the referenced issue is resolved.
	if !rootCmd.PersistentFlags().Lookup("log-format").Changed {
		if lf := logCfg.GetString("format"); lf != "" {
			logFormat = lf
		}
	}
	if !rootCmd.PersistentFlags().Lookup("log-level").Changed {
		if ll := logCfg.GetString("level"); ll != "" {
			logLevel = ll
		}
	}

	if err := log.Init(logLevel, logFormat); err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to initialize logger: %v\n", progName, err)
		os.Exit(1)
	}
}

func InitConfig() {
	// Set defaults for any keys not set by env var, config file, or flag
	config.SetDefaults()

	// Read any environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Do not read or write config file if --ignore-config passed
	if rootCmd.Flag("ignore-config").Changed {
		return
	}

	// Set config file to ~/.config/ochami/config.<configFormat> if not set
	// via flag
	if configFile == "" {
		user, err := user.Current()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: unable to fetch current user: %v\n", progName, err)
			os.Exit(1)
		}
		configDir := filepath.Join(user.HomeDir, ".config", "ochami")
		configFile = filepath.Join(configDir, "config."+configFormat)
	}

	// Try to create config file with default values if it doesn't exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		respConfigCreate := loopYesNo(fmt.Sprintf("Config file %s does not exist. Create it?", configFile))
		if respConfigCreate {
			configDir := filepath.Dir(configFile)
			err := os.MkdirAll(configDir, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: could not create config dir %s: %v\n", progName, configDir, err)
				os.Exit(1)
			}
			f, err := os.OpenFile(configFile, os.O_RDONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: creating %s failed: %v\n", progName, configFile, err)
				os.Exit(1)
			}
			f.Close()
			err = config.WriteConfig(configFile, configFormat)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: writing %s failed: %v\n", progName, configFile, err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "%s: not creating config file. Exiting...\n", progName)
			os.Exit(0)
		}
	}

	// Read configuration file if passed
	err := config.LoadConfig(configFile, configFormat)
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Fprintf(os.Stderr, "%s: configuration file %s not found: %v\n", progName, configFile, err)
		} else {
			fmt.Fprintf(os.Stderr, "%s: failed to load configuration file %s: %v\n", progName, configFile, err)
		}
		os.Exit(1)
	}
}

func InitialStatus() {
	if configFile != "" {
		log.Logger.Debug().Msgf("config file loaded: %s", configFile)
	} else {
		log.Logger.Debug().Msgf("no config file loaded")
	}
}

// prompt displays a text prompt and returns what the user entered. It continues
// to repeat the prompt as long as the user input is empty.
func prompt(prompt string) string {
	var s string
	resp := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, prompt+" ")
		s, _ = resp.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

// loopYesNo takes prompt p and appends " [yN]: " to it and prompts the user for
// input. As long as the user's input is not "y" or "n" (case insensitive), the
// function redisplays the prompt. If the user's response is "y", true is
// returned. If the user's response is "n", false is returned.
func loopYesNo(p string) bool {
	for {
		resp := prompt(fmt.Sprintf("%s [yN]:", p))
		switch strings.ToLower(resp) {
		case "y":
			return true
		case "n":
			return false
		default:
			continue
		}
	}
}

// checkToken takes a pointer to a Cobra command and checks to see if --token
// was set. If not, an error is printed and the program exits.
func checkToken(cmd *cobra.Command) {
	// TODO: Check token validity/expiration
	if token == "" {
		log.Logger.Error().Msg("no token set")
		os.Exit(1)
	}

	// Try to parse token
	t, err := jwt.ParseString(token, jwt.WithValidate(false))
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to parse token")
		os.Exit(1)
	}

	// Check expiration
	now := time.Now()
	exp := t.Expiration()
	if exp.Compare(now) < 0 {
		log.Logger.Error().Msgf("token is expired (expired %s ago at %s)",
			now.Sub(exp), exp.Local().Format(time.RFC1123))
		os.Exit(1)
	} else if exp.Sub(now).Minutes() <= 15 {
		log.Logger.Warn().Msgf("%s until token expires", exp.Sub(now))
	}

	// Validate not before (nbf), issued at (iat), and expiration (exp) fields
	err = jwt.Validate(t,
		jwt.WithValidator(jwt.IsNbfValid()),
		jwt.WithValidator(jwt.IsIssuedAtValid()),
		jwt.WithValidator(jwt.IsExpirationValid()),
	)
	if err != nil {
		log.Logger.Error().Err(err).Msg("token is invalid")
		os.Exit(1)
	}
}

// useCACert takes a pointer to a client.OchamiClient and, if a path to a CA
// certificate has been set via --cacert, it configures it to use it. If an
// error occurs, a log is printed and the program exits.
func useCACert(client *client.OchamiClient) {
	if cacertPath != "" {
		log.Logger.Debug().Msgf("Attempting to use CA certificate at %s", cacertPath)
		if err := client.UseCACert(cacertPath); err != nil {
			log.Logger.Error().Err(err).Msgf("failed to load CA certificate %s", cacertPath)
			os.Exit(1)
		}
	}
}

func getBaseURI(cmd *cobra.Command) (string, error) {
	// Precedence of getting base URI for requests:
	//
	// 1. If --cluster is set, search config file for matching name and read
	//    details from there.
	// 2. If flags corresponding to cluster info (e.g. --base-uri) are set,
	//    read details from them.
	// 3. If "default-cluster" is set in config file (config file must be
	//    specified), use cluster identified by that name as source of info.
	// 4. Data sources exhausted, err.
	var (
		clusterList  []map[string]any
		clusterToUse *map[string]any
		clusterName  string
	)
	if cmd.Flag("cluster").Changed {
		if configFile == "" {
			return "", fmt.Errorf("--cluster specified but no config file specified")
		}
		if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
			return "", fmt.Errorf("failed to unmarshal cluster list: %w", err)
		}
		clusterName = cmd.Flag("cluster").Value.String()
		for _, c := range clusterList {
			if c["name"] == clusterName {
				clusterToUse = &c
				break
			}
		}
		if clusterToUse == nil {
			return "", fmt.Errorf("cluster %q not found in %s", clusterName, configFile)
		}
		clusterToUseData := (*clusterToUse)["cluster"].(map[string]any)
		if clusterToUseData["base-uri"] == nil {
			return "", fmt.Errorf("base-uri not set for cluster %q specified with --cluster", clusterName)
		}
		return clusterToUseData["base-uri"].(string), nil
	} else if cmd.Flag("base-uri").Changed {
		return baseURI, nil
	} else if configFile != "" && viper.IsSet("default-cluster") {
		clusterName = viper.GetString("default-cluster")
		if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
			return "", fmt.Errorf("failed to unmarshal cluster list: %w", err)
		}
		for _, c := range clusterList {
			if c["name"] == clusterName {
				clusterToUse = &c
				break
			}
		}
		if clusterToUse == nil {
			return "", fmt.Errorf("default cluster %q not found in %s", clusterName, configFile)
		}
		clusterToUseData := (*clusterToUse)["cluster"].(map[string]any)
		if clusterToUseData["base-uri"] == nil {
			return "", fmt.Errorf("base-uri not set for default cluster %q", clusterName)
		}
		return clusterToUseData["base-uri"].(string), nil
	}

	return "", fmt.Errorf("no base-uri set bia --base-uri, --cluster, or config file")
}

// setTokenFromEnvVar sets the access token for a cobra command cmd. If --token
// was passed, that value is set as the access token. Otherwise, the token is
// read from an environment variable whose format is <CLUSTER>_ACCESS_TOKEN
// where <CLUSTER> is the name of the cluster, in upper case, being contacted.
// The value of <CLUSTER> is determined by taking the cluster name, passed
// either by --cluster or reading default-cluster from the config file (the
// former preceding the latter), replacing spaces and dashes (-) with
// underscores, and making the letters uppercase. If no config file is set or
// the environment variable is not set, an error is logged and the program
// exits.
func setTokenFromEnvVar(cmd *cobra.Command) {
	var (
		clusterName string
		varPrefix   string
	)
	if cmd.Flag("token").Changed {
		token = cmd.Flag("token").Value.String()
		log.Logger.Debug().Msg("--token passed, setting token to its value: " + token)
		return
	} else if configFile != "" {
		log.Logger.Debug().Msg("Determining token from environment variable based on cluster in config file")
		if cmd.Flag("cluster").Changed {
			clusterName = cmd.Flag("cluster").Value.String()
			log.Logger.Debug().Msg("--cluster specified: " + clusterName)
		} else if viper.IsSet("default-cluster") {
			clusterName = viper.GetString("default-cluster")
			log.Logger.Debug().Msg("--cluster not specified, using default-cluster: " + clusterName)
		}
	} else {
		log.Logger.Error().Msg("no config file specified to determine which cluster token to use and --token not specified")
		os.Exit(1)
	}
	varPrefix = strings.ReplaceAll(clusterName, "-", "_")
	varPrefix = strings.ReplaceAll(varPrefix, " ", "_")

	envVarToRead := strings.ToUpper(varPrefix) + "_ACCESS_TOKEN"
	log.Logger.Debug().Msg("Reading token from environment variable: " + envVarToRead)
	if t, tokenSet := os.LookupEnv(envVarToRead); tokenSet {
		log.Logger.Debug().Msgf("Token found from environment variable: %s=%s", envVarToRead, t)
		token = t
		return
	}

	log.Logger.Error().Msgf("Environment variable %s unset for reading token for cluster %q", envVarToRead, clusterName)
	os.Exit(1)
}

// handlePayload unmarshals a payload file into data for command cmd if
// --payload and, optionally, --payload-format, are passed.
func handlePayload(cmd *cobra.Command, data any) {
	if cmd.Flag("payload").Changed {
		dFile := cmd.Flag("payload").Value.String()
		dFormat := cmd.Flag("payload-format").Value.String()
		err := client.ReadPayload(dFile, dFormat, data)
		if err != nil {
			log.Logger.Error().Err(err).Msg("unable to read payload for request")
			os.Exit(1)
		}
	}
}
