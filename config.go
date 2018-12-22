package main

import (
	"reflect"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/spf13/viper"
)

type gridColumnConfig struct {
	key    string
	header string
	width  int
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

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.lncli-curses")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
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
