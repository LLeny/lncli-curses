package main

import (
	"fmt"
	"reflect"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/jroimartin/gocui"
	"github.com/spf13/viper"
)

type cliOpts struct {
	LncliExec       string `short:"l" long:"lnclicmd" description:"lncli executable"`
	RefreshSec      int    `short:"r" long:"refresh" description:"lncli data refresh time in seconds"`
	RPCServer       string `long:"rpcserver" description:"host:port of ln daemon"`
	LndDir          string `long:"lnddir" description:"path to lnd's base directory"`
	TLSCertPath     string `long:"tlscertpath" description:"path to TLS certificate"`
	NoMacaroons     bool   `long:"no-macaroons" description:"disable macaroon authentication"`
	MacaroonPath    string `long:"macaroonpath" description:"path to macaroon file"`
	MacaroonTimeOut int    `long:"macaroontimeout" description:"anti-replay macaroon validity time in seconds"`
	MacaroonIP      string `long:"macaroonip" description:"if set, lock macaroon to specific IP address"`
}

var (
	cfgShowHeader bool
	cfgOpts       cliOpts
)

type gridColumnConfig struct {
	key    string
	header string
	width  int
}

func getLncliExec() string {
	if len(cfgOpts.LncliExec) > 0 {
		return cfgOpts.LncliExec
	}
	return "lncli"
}

func getRefreshSec() int {
	if cfgOpts.RefreshSec > 0 {
		return cfgOpts.RefreshSec
	}
	return 60
}

func getRPCServer() string {
	return cfgOpts.RPCServer
}

func getLndDir() string {
	return cfgOpts.LndDir
}

func getTLSCertPath() string {
	return cfgOpts.TLSCertPath
}

func getNoMacaroons() bool {
	return cfgOpts.NoMacaroons
}

func getMacaroonPath() string {
	return cfgOpts.MacaroonPath
}

func getMacaroonTimeOut() int {
	return cfgOpts.MacaroonTimeOut
}

func getMacaroonIP() string {
	return cfgOpts.MacaroonIP
}

func getShowHeader() bool {
	return cfgShowHeader
}

func getConfigString(key string) string {
	return viper.GetString(key)
}

func getConfigInt(key string) int {
	return viper.GetInt(key)
}

func getThemeBashColor(key string) string {
	return strings.Replace(getConfigString(key), "[", "\x1b[", -1)
}

func getConfigGridColumns(gridKey string) []gridColumnConfig {
	grid := viper.GetStringMap("grids." + gridKey)

	if grid == nil {
		panic("Grid configuration not found")
	}

	s := reflect.ValueOf(grid["columns"])
	if s.Kind() != reflect.Slice {
		panic("Grid configuration invalid")
	}

	ret := make([]gridColumnConfig, s.Len())

	for i := 0; i < s.Len(); i++ {
		col := s.Index(i).Interface().(map[string]interface{})
		ret[i] = gridColumnConfig{col["key"].(string), col["header"].(string), int(col["width"].(float64))}
	}

	return ret
}

func getConfigGridShortcutHeader(gridKey string) string {
	return viper.GetString("grids." + gridKey + ".shortcutHeader")
}

func getConfigGridHeader(gridKey string) string {
	return viper.GetString("grids." + gridKey + ".header")
}

func initConfig() bool {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.lncli-curses")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	initBaseConfig()

	if !readCommandLine() {
		return false
	}

	return true
}

func readCommandLine() bool {

	var opts cliOpts

	if _, err := flags.Parse(&opts); err != nil {
		fmt.Println(err)
		return false
	}

	if len(opts.LncliExec) > 0 {
		cfgOpts.LncliExec = opts.LncliExec
	}

	if len(opts.MacaroonIP) > 0 {
		cfgOpts.MacaroonIP = opts.MacaroonIP
	}

	if len(opts.MacaroonPath) > 0 {
		cfgOpts.MacaroonPath = opts.MacaroonPath
	}

	if opts.MacaroonTimeOut > 0 {
		cfgOpts.MacaroonTimeOut = opts.MacaroonTimeOut
	}

	if opts.NoMacaroons {
		cfgOpts.NoMacaroons = opts.NoMacaroons
	}

	if len(opts.RPCServer) > 0 {
		cfgOpts.RPCServer = opts.RPCServer
	}

	if opts.RefreshSec > 0 {
		cfgOpts.RefreshSec = opts.RefreshSec
	}

	if len(opts.TLSCertPath) > 0 {
		cfgOpts.TLSCertPath = opts.TLSCertPath
	}

	if len(opts.LndDir) > 0 {
		cfgOpts.LndDir = opts.LndDir
	}

	return true
}

func initBaseConfig() {
	cfgShowHeader = viper.GetBool("showheader")
	cfgOpts.LncliExec = viper.GetString("LncliExec")
	cfgOpts.RefreshSec = viper.GetInt("RefreshSec")
	cfgOpts.RPCServer = viper.GetString("RPCServer")
	cfgOpts.LndDir = viper.GetString("LndDir")
	cfgOpts.TLSCertPath = viper.GetString("TLSCertPath")
	cfgOpts.NoMacaroons = viper.GetBool("NoMacaroons")
	cfgOpts.MacaroonPath = viper.GetString("MacaroonPath")
	cfgOpts.MacaroonTimeOut = viper.GetInt("MacaroonTimeOut")
	cfgOpts.MacaroonIP = viper.GetString("MacaroonIP")
}

func initTheme() {
	context.theme.background = gocui.Attribute(getConfigInt("theme.background"))
	context.theme.inverted = getThemeBashColor("theme.inverted")
	context.theme.highlight = getThemeBashColor("theme.highlight")
	context.theme.error = getThemeBashColor("theme.error")
	context.theme.labelHeader = getThemeBashColor("theme.labelHeader")
	context.theme.normal = getThemeBashColor("theme.normal")
	context.theme.bold = getThemeBashColor("theme.bold")
	context.theme.gridHeader = getThemeBashColor("theme.gridHeader")
	context.theme.gridSelected = getThemeBashColor("theme.gridSelected")
}
