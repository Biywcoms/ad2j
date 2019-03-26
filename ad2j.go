package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
	"utils"
)

// UTF16BytesToString converts UTF-16 encoded bytes, in big or little endian byte order,
// to a UTF-8 encoded string.
func UTF16BytesToString(b []byte, o binary.ByteOrder) string {
	utf := make([]uint16, (len(b)+(2-1))/2)
	for i := 0; i+(2-1) < len(b); i += 2 {
		utf[i/2] = o.Uint16(b[i:])
	}
	if len(b)/2 < len(utf) {
		utf[len(utf)-1] = utf8.RuneError
	}
	return string(utf16.Decode(utf))
}

// UTF-16 endian byte order
const (
	unknownEndian = iota
	bigEndian
	littleEndian
)

// dropCREndian drops a terminal \r from the endian data.
func dropCREndian(data []byte, t1, t2 byte) []byte {
	if len(data) > 1 {
		if data[len(data)-2] == t1 && data[len(data)-1] == t2 {
			return data[0 : len(data)-2]
		}
	}
	return data
}

// dropCRBE drops a terminal \r from the big endian data.
func dropCRBE(data []byte) []byte {
	return dropCREndian(data, '\x00', '\r')
}

// dropCRLE drops a terminal \r from the little endian data.
func dropCRLE(data []byte) []byte {
	return dropCREndian(data, '\r', '\x00')
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) ([]byte, int) {
	var endian = unknownEndian
	switch ld := len(data); {
	case ld != len(dropCRLE(data)):
		endian = littleEndian
	case ld != len(dropCRBE(data)):
		endian = bigEndian
	}
	return data, endian
}

// SplitFunc is a split function for a Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanUTF16LinesFunc(byteOrder binary.ByteOrder) (bufio.SplitFunc, func() binary.ByteOrder) {

	// Function closure variables
	var endian = unknownEndian
	switch byteOrder {
	case binary.BigEndian:
		endian = bigEndian
	case binary.LittleEndian:
		endian = littleEndian
	}
	const bom = 0xFEFF
	var checkBOM bool = endian == unknownEndian

	// Scanner split function
	splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if checkBOM {
			checkBOM = false
			if len(data) > 1 {
				switch uint16(bom) {
				case uint16(data[0])<<8 | uint16(data[1]):
					endian = bigEndian
					return 2, nil, nil
				case uint16(data[1])<<8 | uint16(data[0]):
					endian = littleEndian
					return 2, nil, nil
				}
			}
		}

		// Scan for newline-terminated lines.
		i := 0
		for {
			j := bytes.IndexByte(data[i:], '\n')
			if j < 0 {
				break
			}
			i += j
			switch e := i % 2; e {
			case 1: // UTF-16BE
				if endian != littleEndian {
					if i > 1 {
						if data[i-1] == '\x00' {
							endian = bigEndian
							// We have a full newline-terminated line.
							return i + 1, dropCRBE(data[0 : i-1]), nil
						}
					}
				}
			case 0: // UTF-16LE
				if endian != bigEndian {
					if i+1 < len(data) {
						i++
						if data[i] == '\x00' {
							endian = littleEndian
							// We have a full newline-terminated line.
							return i + 1, dropCRLE(data[0 : i-1]), nil
						}
					}
				}
			}
			i++
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			// drop CR.
			advance = len(data)
			switch endian {
			case bigEndian:
				data = dropCRBE(data)
			case littleEndian:
				data = dropCRLE(data)
			default:
				data, endian = dropCR(data)
			}
			if endian == unknownEndian {
				if runtime.GOOS == "windows" {
					endian = littleEndian
				} else {
					endian = bigEndian
				}
			}
			return advance, data, nil
		}

		// Request more data.
		return 0, nil, nil
	}

	// Endian byte order function
	orderFunc := func() (byteOrder binary.ByteOrder) {
		switch endian {
		case bigEndian:
			byteOrder = binary.BigEndian
		case littleEndian:
			byteOrder = binary.LittleEndian
		}
		return byteOrder
	}

	return splitFunc, orderFunc
}

func Revenue(lst [][]string, m, n int) (result [][]string) {
	for _, line := range lst {
		vm, err1 := strconv.ParseFloat(line[m], 32)
		vn, err2 := strconv.ParseFloat(line[n], 32)

		var v string
		if err1 != nil || err2 != nil {
			v = "0"
		} else {
			v = strconv.FormatFloat(vm*vn, 'f', -1, 32)
		}

		line = append(line, v)
		result = append(result, line)
	}
	return result

}

var filename string

var COUNTRY = map[string]string{
	"美国":       "us",
	"伊朗":       "ir",
	"越南":       "vn",
	"波兰":       "pl",
	"法国":       "fr",
	"台湾":       "tw",
	"巴西":       "br",
	"泰国":       "th",
	"德国":       "de",
	"英国":       "gb",
	"印度":       "in",
	"秘鲁":       "pe",
	"日本":       "jp",
	"埃及":       "eg",
	"希腊":       "gr",
	"中国":       "cn",
	"香港":       "hk",
	"瑞士":       "ch",
	"捷克":       "sz",
	"荷兰":       "nl",
	"缅甸":       "mm",
	"挪威":       "no",
	"瑞典":       "se",
	"智利":       "cl",
	"南非":       "za",
	"韩国":       "kr",
	"文莱":       "bn",
	"芬兰":       "fi",
	"奥地利":      "at",
	"比利时":      "be",
	"巴拿马":      "pa",
	"以色列":      "il",
	"土耳其":      "tr",
	"墨西哥":      "mx",
	"西班牙":      "es",
	"葡萄牙":      "pt",
	"新西兰":      "nz",
	"俄罗斯":      "ru",
	"意大利":      "it",
	"乌克兰":      "ua",
	"摩洛哥":      "ma",
	"阿联酋":      "ae",
	"菲律宾":      "ph",
	"阿根廷":      "ar",
	"伊拉克":      "iq",
	"新加坡":      "sg",
	"加拿大":      "ca",
	"匈牙利":      "hu",
	"科索沃":      "xk",
	"斯洛伐克":     "sk",
	"澳大利亚":     "au",
	"拉脱维亚":     "lv",
	"白俄罗斯":     "by",
	"巴基斯坦":     "pk",
	"马来西亚":     "my",
	"哥伦比亚":     "co",
	"罗马尼亚":     "ro",
	"孟加拉国":     "bd",
	"格鲁吉亚":     "ge",
	"危地马拉":     "gt",
	"哥斯达黎加":    "cr",
	"印度尼西亚":    "id",
	"哈萨克斯坦":    "kz",
	"沙特阿拉伯":    "sa",
	"阿尔及利亚":    "dz",
	"阿尔巴尼亚":    "al",
	"斯洛文尼亚":    "si",
	"新喀里多尼亚":   "nc",
	"阿拉伯联合酋长国": "ae",
}

type Apps struct {
	Market string   `json:"market"`
	App    []Detail `json:"app"`
}

type Detail struct {
	Country []string `json:"country"`
	Pkg     string   `json:"pkg"`
	Url     string   `json:"url"`
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	fullfilename := path.Base(file.Name())
	fileSuffix := path.Ext(fullfilename)
	lastname := strings.Split(fullfilename, "-")
	filename = strings.TrimSuffix(lastname[len(lastname)-1], fileSuffix)

	rdr := bufio.NewReader(file)
	scanner := bufio.NewScanner(rdr)
	var bo binary.ByteOrder // unknown, infer from data
	// bo = binary.LittleEndian // windows
	splitFunc, orderFunc := ScanUTF16LinesFunc(bo)
	scanner.Split(splitFunc)

	var data [][]string

	for scanner.Scan() {
		b := scanner.Bytes()
		s := UTF16BytesToString(b, orderFunc())
		lst := strings.Split(s, "\t")
		data = append(data, []string{lst[0], lst[1], lst[12], lst[13]})
		//		fmt.Println(len(s), s)
		//		fmt.Println(len(b), b)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	result := utils.ArraySort(Revenue(data[1:], 2, 3), 4, true)

	//	fmt.Println(orderFunc())
	m := make(map[string]string)
	for _, n := range result {
		m[n[0]] = n[0]
		if len(m) >= 5 {
			break
		}
	}

	var apps []Detail

	for _, n := range m {
		var i int
		var countryCode []string
		for _, line := range result {
			if n == line[0] {
				elem, ok := COUNTRY[line[1]]
				if ok {
					line[1] = elem
				}
				countryCode = append(countryCode, line[1])
				//				edata = append(edata, line)
				i++
			}
			if i >= 10 {
				break
			}
		}
		var detail Detail
		detail.Country = countryCode
		detail.Pkg = n
		detail.Url = "https://lh3.ggpht.com/=s180"
		apps = append(apps, detail)
	}

	var mapps Apps
	mapps.Market = "https://play.google.com/store/apps/details?id="
	mapps.App = apps

	b, err := json.MarshalIndent(mapps, "", " ")

	if err != nil {
		panic(err)
		return
	}

	utils.ExportFileS("/home/jake/Web/", "app-"+filename+".json", string(b))

}
