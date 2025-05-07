package configreader

import (
	"encoding"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"

	"fknsrs.biz/p/ytmusic/internal/stringutil"
)

func Read(program string, arguments, environment []string, out interface{}) error {
	if _, _, err := getValueAndType(out); err != nil {
		return fmt.Errorf("configreader.Read: could not get value and type: %w", err)
	}

	if configPath, ok := getFromArgumentsOrEnvironmentOrObject(arguments, environment, out, "config"); ok && configPath != "" {
		if err := readFile(configPath, out); err != nil {
			return fmt.Errorf("configreader.Read: %w", err)
		}
	}

	if err := readArguments(program, arguments, out); err != nil {
		return fmt.Errorf("configreader.Read: could not read command-line flags: %w", err)
	}

	if err := readEnvironment(program, environment, out); err != nil {
		return fmt.Errorf("configreader.Read: could not read environment variables: %w", err)
	}

	return nil
}

func getValueAndType(v interface{}) (reflect.Value, reflect.Type, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, nil, fmt.Errorf("configreader.getValueAndType: value must be a non-nil pointer; was instead %T", v)
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("configreader.getValueAndType: value must be a pointer to a struct; was instead %T", v)
	}

	return rv, rv.Type(), nil
}

type encodingText interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

var (
	stringType       = reflect.TypeOf("")
	boolType         = reflect.TypeOf(true)
	intType          = reflect.TypeOf(int(0))
	encodingTextType = reflect.TypeOf((*encodingText)(nil)).Elem()
)

func getFromArgumentsOrEnvironmentOrObject(arguments, environment []string, obj interface{}, name string) (string, bool) {
	if s, ok := getFromArguments(arguments, name); ok {
		return s, ok
	}

	if s, ok := getFromEnvironment(environment, name); ok {
		return s, ok
	}

	if s, ok := getFromObject(obj, name); ok {
		return s, ok
	}

	return "", false
}

func getFromArguments(arguments []string, name string) (string, bool) {
	prefix := "-" + name

	for i := 0; i < len(arguments); i++ {
		if arguments[i] == prefix && i+1 < len(arguments) {
			return arguments[i+1], true
		} else if strings.HasPrefix(arguments[i], prefix+"=") {
			return strings.TrimPrefix(arguments[i], prefix+"="), true
		}
	}

	return "", false
}

func getFromEnvironment(environment []string, name string) (string, bool) {
	prefix := strings.ToLower(name + "=")

	for i := 0; i < len(environment); i++ {
		if strings.HasPrefix(strings.ToLower(environment[i]), prefix) {
			return environment[i][len(prefix):], true
		}
	}

	return "", false
}

func getFromObject(obj interface{}, name string) (string, bool) {
	val, typ, err := getValueAndType(obj)
	if err != nil {
		return "", false
	}

	for i := 0; i < val.NumField(); i++ {
		vf := val.Field(i)
		tf := typ.Field(i)

		parameterName, _, ok := getNameAndHelpForField(tf)
		if !ok {
			continue
		}

		if parameterName != name {
			continue
		}

		switch {
		case tf.Type == stringType:
			return vf.String(), true
		case reflect.PointerTo(tf.Type).Implements(encodingTextType) && vf.IsValid():
			d, err := vf.Addr().Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return "", false
			}
			return string(d), true
		}
	}

	return "", false
}

func readFile(filePath string, out interface{}) error {
	switch filepath.Ext(filePath) {
	case ".yaml", ".yml":
		if err := readFileYAML(filePath, out); err != nil {
			return fmt.Errorf("readFile: could not read %q as yaml: %w", filePath, err)
		}

		return nil
	case ".toml":
		if err := readFileTOML(filePath, out); err != nil {
			return fmt.Errorf("readFile: could not read %q as toml: %w", filePath, err)
		}

		return nil
	default:
		return fmt.Errorf("readFile: could not determine file type for %q", filePath)
	}
}

func readFileYAML(filePath string, out interface{}) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("readFileYAML: could not open config file: %w", err)
	}
	defer fd.Close()

	if err := yaml.NewDecoder(fd).Decode(out); err != nil {
		return fmt.Errorf("readFileYAML: could not parse config file: %w", err)
	}

	return nil
}

func readFileTOML(filePath string, out interface{}) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("readFileTOML: could not open config file: %w", err)
	}
	defer fd.Close()

	if err := toml.NewDecoder(fd).Decode(out); err != nil {
		return fmt.Errorf("readFileTOML: could not parse config file: %w", err)
	}

	return nil
}

func readArguments(program string, arguments []string, out interface{}) error {
	val, typ, err := getValueAndType(out)
	if err != nil {
		return fmt.Errorf("configreader.readArguments: could not get value and type: %w", err)
	}

	flagSet := flag.NewFlagSet(program, flag.ContinueOnError)

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", program)
		flagSet.PrintDefaults()
		os.Exit(0)
	}

	for i := 0; i < val.NumField(); i++ {
		vf := val.Field(i)
		tf := typ.Field(i)

		name, help, ok := getNameAndHelpForField(tf)
		if !ok {
			continue
		}

		switch {
		case tf.Type == stringType:
			flagSet.StringVar(vf.Addr().Interface().(*string), name, vf.String(), help)
		case tf.Type == boolType:
			flagSet.BoolVar(vf.Addr().Interface().(*bool), name, vf.Bool(), help)
		case tf.Type == intType:
			flagSet.IntVar(vf.Addr().Interface().(*int), name, int(vf.Int()), help)
		case reflect.PointerTo(tf.Type).Implements(encodingTextType):
			flagSet.TextVar(vf.Addr().Interface().(encoding.TextUnmarshaler), name, vf.Addr().Interface().(encoding.TextMarshaler), help)
		default:
			return fmt.Errorf("configreader.readArguments: could not define flag for parameter %s (%s) with type %s", tf.Name, name, tf.Type)
		}
	}

	if err := flagSet.Parse(arguments); err != nil {
		return err
	}

	return nil
}

func readEnvironment(program string, environment []string, out interface{}) error {
	val, typ, err := getValueAndType(out)
	if err != nil {
		return fmt.Errorf("configreader.readEnvironment: could not get value and type: %w", err)
	}

	for i := 0; i < typ.NumField(); i++ {
		vf := val.Field(i)
		tf := typ.Field(i)

		name, _, ok := getNameAndHelpForField(tf)
		if !ok {
			continue
		}

		ev, ok := getFromEnvironment(environment, name)
		if !ok {
			continue
		}

		switch {
		case tf.Type == stringType:
			vf.SetString(ev)
		case reflect.PointerTo(tf.Type).Implements(encodingTextType):
			if err := vf.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(ev)); err != nil {
				return fmt.Errorf("configreader.readEnvironment: could not unmarshal parameter %s (%s): %w", tf.Name, name, err)
			}
		default:
			return fmt.Errorf("configreader.readEnvironment: could not read parameter %s (%s) of type %s", tf.Name, name, tf.Type)
		}
	}

	return nil
}

func getNameAndHelpForField(f reflect.StructField) (string, string, bool) {
	name := f.Tag.Get("name")
	if name == "" {
		name = stringutil.PascalToSnake(f.Name)
	}

	help := f.Tag.Get("help")

	switch name {
	case "-":
		return "", "", false
	default:
		return name, help, true
	}
}
