// toolfunc project go
package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/zlib"
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"isotime"
	"log"
	"math"
	"math/rand"
	"method"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/nfnt/resize"

	"netutil"

	"github.com/cypro666/mahonia"
)

var istoolfuncinit bool

func InitToolFunc() {

	istoolfuncinit = true
}

func OrderFileClear(path string) {
	os.Remove(path)
	os.Remove(path + ".head")
	os.Remove(path + ".head.backup")
	os.Remove(path + ".free")
	os.Remove(path + ".free.backup")
	os.Remove(path + ".rmpush")
	os.Remove(path + ".rmpushinsave")
	os.Remove(path + ".newsavedata")
	os.Remove(path + ".headsaveok")
	os.Remove(path + ".rmpushinmem")
	os.Remove(path + ".headdb")
	os.Remove(path + ".headdb.head")
	os.Remove(path + ".headdb.head.backup")
	os.Remove(path + ".headdb.free")
	os.Remove(path + ".headdb.free.backup")
	os.Remove(path + ".headdb.rmpush")
	os.Remove(path + ".headdb.newsavedata")
	os.Remove(path + ".headdb.rmpushinsave")
	os.Remove(path + ".headdb.rmpushinmem")
	os.Remove(path + ".headdb.headsaveok")
}

func GetGoRoutineID() int {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic recover:panic info:%V", err)
		}
	}()

	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func CopyFile(src, target string) bool {
	target = strings.Replace(target, "\\", "/", -1)
	oldrmpushf, oldrmpushferr := os.OpenFile(src, os.O_RDWR, 0666)
	if oldrmpushferr == nil {
		if strings.LastIndex(target, "/") != -1 {
			targetdir := target[:strings.LastIndex(target, "/")]
			os.MkdirAll(targetdir, 0666)
		}
		compactrmpushf, _ := os.Create(target)
		endpos, _ := oldrmpushf.Seek(0, os.SEEK_END)
		if endpos > 0 {
			readbuf := make([]byte, 2*1024*1024)
			oldrmpushf.Seek(0, os.SEEK_SET)
			for i := int64(0); i < endpos; {
				readlen := int64(len(readbuf))
				if i+readlen > endpos {
					readlen = endpos - i
				}
				_, readerr := oldrmpushf.Read(readbuf[:readlen])
				if readerr != nil {
					compactrmpushf.Close()
					oldrmpushf.Close()
					return false
				}
				compactrmpushf.Write(readbuf[:readlen])
				i += readlen
			}
		}
		compactrmpushf.Close()
		oldrmpushf.Close()
	} else {
		return false
	}
	return true
}

func CopyDir(src, target string) bool {
	src = StdDir(src)
	d, err := os.Open(src)
	if err != nil {
		return false
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return false
	}
	for _, name := range names {
		info, err := os.Stat(src + name)
		if err == nil {
			if info.IsDir() {
				CopyDir(src+name, target+name)
			} else {
				CopyFile(src+name, target+name)
			}
		} else {
			return false
		}
	}
	return true
}

//add dir suffix string
func FilePathAppendDir(filepath, dirname string) string {
	filepath = StdPath(filepath)
	if strings.LastIndex(filepath, "/") != -1 {
		return filepath[:strings.LastIndex(filepath, "/")+1] + dirname + "/" + filepath[strings.LastIndex(filepath, "/")+1:]
	} else {
		return dirname + "/" + filepath
	}
}

//add dir suffix string
func FilePathAddDirSuffix(filepath, dirname string) string {
	filepath = StdPath(filepath)
	if strings.LastIndex(filepath, "/") != -1 {
		return filepath[:strings.LastIndex(filepath, "/")] + dirname + "/" + filepath[strings.LastIndex(filepath, "/")+1:]
	} else {
		return dirname + "/" + filepath
	}
}

func FilePathReplaceDir(filepath, dirname string) string {
	filepath = StdPath(filepath)
	filename := filepath[strings.LastIndex(filepath, "/")+1:]
	filepath = filepath[:strings.LastIndex(filepath, "/")]
	return filepath[:strings.LastIndex(filepath, "/")+1] + dirname + "/" + filename
}

func FileSize(filepath string) int64 {
	dl, dler := os.Stat(filepath)
	if dler == nil {
		if !dl.IsDir() {
			ff, ffer := os.OpenFile(filepath, os.O_RDONLY, 0666)
			if ffer == nil {
				endpos, sker := ff.Seek(0, os.SEEK_END)
				ff.Close()
				if sker == nil {
					return endpos
				}
			}
		}
	}
	return -1
}

func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func BytesJoin(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

func EmptifyDir(dir string) error {
	if dir[len(dir)-1:] != "\\" {
		dir += "\\"
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Println("test remove:", dir+name)
		err = os.RemoveAll(dir + name)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveDirAll(dir string) error {
	return os.RemoveAll(dir)
}

func IsDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

func IsFileExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return !fi.IsDir()
	}
}

func DeleteSubDir(parentdir string) error {
	if parentdir[len(parentdir)-1:] != "\\" {
		parentdir += "\\"
	}
	d, err := os.Open(parentdir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		stat, staterr := os.Stat(parentdir + name)
		if staterr == nil && stat.IsDir() {
			fmt.Println("remove sub dir:", parentdir+name)
			err = os.RemoveAll(parentdir + name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DeleteSubFile(parentdir string) error {
	if parentdir[len(parentdir)-1:] != "/" {
		parentdir += "/"
	}
	d, err := os.Open(parentdir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		stat, staterr := os.Stat(parentdir + name)
		if staterr == nil && !stat.IsDir() {
			err = os.Remove(parentdir + name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func U8FirstUChar(data []byte) (unichar uint32, unicharoplen int, success bool) {
	utf8_byte_mask := uint32(0x3f)
	datai := 0
	for datai < len(data) {
		lead := data[datai]
		// 0xxxxxxx -> U+0000..U+007F
		if lead < 0x80 {
			unichar = uint32(lead)
			if unichar < 0x80 {
				unicharoplen = 1
			} else if unichar < 0x800 {
				unicharoplen = 2
			} else {
				unicharoplen = 3
			}
			return unichar, unicharoplen, true
		} else if (lead-0xC0) < 0x20 && len(data) >= 2 && (data[1]&0xc0) == 0x80 {
			// 110xxxxx -> U+0080..U+07FF
			unichar = ((uint32(lead) & ^uint32(0xC0)) << 6) | (uint32(data[1]) & utf8_byte_mask)
			if unichar < 0x80 {
				unicharoplen = 1
			} else if unichar < 0x800 {
				unicharoplen = 2
			} else {
				unicharoplen = 3
			}
			return unichar, unicharoplen, true
		} else if (lead-0xE0) < 0x10 && len(data) >= 3 && (data[1]&0xc0) == 0x80 && (data[2]&0xc0) == 0x80 {
			// 1110xxxx -> U+0800-U+FFFF
			unichar = ((uint32(lead) & ^uint32(0xE0)) << 12) | ((uint32(data[1]) & utf8_byte_mask) << 6) | (uint32(data[2]) & utf8_byte_mask)
			if unichar < 0x80 {
				unicharoplen = 1
			} else if unichar < 0x800 {
				unicharoplen = 2
			} else {
				unicharoplen = 3
			}
			return unichar, unicharoplen, true
		} else if (lead-0xF0) < 0x08 && len(data) >= 4 && (data[1]&0xc0) == 0x80 && (data[2]&0xc0) == 0x80 && (data[3]&0xc0) == 0x80 {
			// 11110xxx -> U+10000..U+10FFFF
			unichar = ((uint32(lead) & ^uint32(0xF0)) << 18) | ((uint32(data[1]) & utf8_byte_mask) << 12) | ((uint32(data[2]) & utf8_byte_mask) << 6) | (uint32(data[3]) & utf8_byte_mask)
			unicharoplen = 4
			return unichar, unicharoplen, true
		} else {
			// 10xxxxxx or 11111xxx -> invalid
			datai += 1
		}
	}
	return 0, 0, false
}

func Contain(set interface{}, val interface{}) bool {
	switch set.(type) {
	case []int:
		{
			for _, val2 := range set.([]int) {
				if val.(int) == val2 {
					return true
				}
			}
		}
	case []int8:
		{
			for _, val2 := range set.([]int8) {
				if val.(int8) == val2 {
					return true
				}
			}
		}
	case []int16:
		{
			for _, val2 := range set.([]int16) {
				if val.(int16) == val2 {
					return true
				}
			}
		}
	case []int32:
		{
			for _, val2 := range set.([]int32) {
				if val.(int32) == val2 {
					return true
				}
			}
		}
	case []int64:
		{
			for _, val2 := range set.([]int64) {
				if val.(int64) == val2 {
					return true
				}
			}
		}
	case []uint8:
		{
			for _, val2 := range set.([]uint8) {
				if val.(uint8) == val2 {
					return true
				}
			}
		}
	case []uint16:
		{
			for _, val2 := range set.([]uint16) {
				if val.(uint16) == val2 {
					return true
				}
			}

		}
	case []uint32:
		{
			for _, val2 := range set.([]uint32) {
				if val.(uint32) == val2 {
					return true
				}
			}

		}
	case []uint64:
		{
			for _, val2 := range set.([]uint64) {
				if val.(uint64) == val2 {
					return true
				}
			}
		}
	case []string:
		{
			for _, val2 := range set.([]string) {
				if val.(string) == val2 {
					return true
				}
			}
		}
	}
	return false
}

func SliceSearch(set interface{}, val interface{}, from int) int {
	switch set.(type) {
	case []int:
		{
			for i := from; i < len(set.([]int)); i++ {
				if val.(int) == set.([]int)[i] {
					return i
				}
			}
		}
	case []int8:
		{
			for i := from; i < len(set.([]int8)); i++ {
				if val.(int8) == set.([]int8)[i] {
					return i
				}
			}
		}
	case []int16:
		{
			for i := from; i < len(set.([]int16)); i++ {
				if val.(int16) == set.([]int16)[i] {
					return i
				}
			}
		}
	case []int32:
		{
			for i := from; i < len(set.([]int32)); i++ {
				if val.(int32) == set.([]int32)[i] {
					return i
				}
			}
		}
	case []int64:
		{
			for i := from; i < len(set.([]int64)); i++ {
				if val.(int64) == set.([]int64)[i] {
					return i
				}
			}
		}
	case []uint8:
		{
			for i := from; i < len(set.([]uint8)); i++ {
				if val.(uint8) == set.([]uint8)[i] {
					return i
				}
			}
		}
	case []uint16:
		{
			for i := from; i < len(set.([]uint16)); i++ {
				if val.(uint16) == set.([]uint16)[i] {
					return i
				}
			}
		}
	case []uint32:
		{
			for i := from; i < len(set.([]uint32)); i++ {
				if val.(uint32) == set.([]uint32)[i] {
					return i
				}
			}
		}
	case []uint64:
		{
			for i := from; i < len(set.([]uint64)); i++ {
				if val.(uint64) == set.([]uint64)[i] {
					return i
				}
			}
		}
	case []string:
		{
			for i := from; i < len(set.([]string)); i++ {
				if val.(string) == set.([]string)[i] {
					return i
				}
			}
		}
	case [][]byte:
		{
			for i := from; i < len(set.([][]byte)); i++ {
				if bytes.Compare(val.([]byte), set.([][]byte)[i]) == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func SliceLastIndex(set interface{}, val interface{}) int {
	switch set.(type) {
	case []int:
		{
			for i := len(set.([]int)) - 1; i >= 0; i-- {
				if val.(int) == set.([]int)[i] {
					return i
				}
			}
		}
	case []int8:
		{
			for i := len(set.([]int8)) - 1; i >= 0; i-- {
				if val.(int8) == set.([]int8)[i] {
					return i
				}
			}
		}
	case []int16:
		{
			for i := len(set.([]int16)) - 1; i >= 0; i-- {
				if val.(int16) == set.([]int16)[i] {
					return i
				}
			}
		}
	case []int32:
		{
			for i := len(set.([]int32)) - 1; i >= 0; i-- {
				if val.(int32) == set.([]int32)[i] {
					return i
				}
			}
		}
	case []int64:
		{
			for i := len(set.([]int64)) - 1; i >= 0; i-- {
				if val.(int64) == set.([]int64)[i] {
					return i
				}
			}
		}
	case []uint8:
		{
			for i := len(set.([]uint8)) - 1; i >= 0; i-- {
				if val.(uint8) == set.([]uint8)[i] {
					return i
				}
			}
		}
	case []uint16:
		{
			for i := len(set.([]uint16)) - 1; i >= 0; i-- {
				if val.(uint16) == set.([]uint16)[i] {
					return i
				}
			}
		}
	case []uint32:
		{
			for i := len(set.([]uint32)) - 1; i >= 0; i-- {
				if val.(uint32) == set.([]uint32)[i] {
					return i
				}
			}
		}
	case []uint64:
		{
			for i := len(set.([]uint64)) - 1; i >= 0; i-- {
				if val.(uint64) == set.([]uint64)[i] {
					return i
				}
			}
		}
	case []string:
		{
			for i := len(set.([]string)) - 1; i >= 0; i-- {
				if val.(string) == set.([]string)[i] {
					return i
				}
			}
		}
	case [][]byte:
		{
			for i := len(set.([][]byte)) - 1; i >= 0; i-- {
				if bytes.Compare(val.([]byte), set.([][]byte)[i]) == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func BytesClone(src []byte) []byte {
	target := make([]byte, len(src))
	copy(target, src)
	return target
}

func StringListCompare(ar1, ar2 []string) int {
	var ln1, ln2, ln int
	ln1 = len(ar1)
	ln2 = len(ar2)
	if ln1 > ln2 {
		ln = ln2
	} else {
		ln = ln1
	}
	for i := 0; i < ln; i++ {
		cmprl := strings.Compare(ar1[i], ar2[i])
		if cmprl != 0 {
			return cmprl
		}
	}
	if ln1 == ln2 {
		return 0
	} else if ln1 > ln2 {
		return 1
	} else {
		return -1
	}
}

func Uint64ListCompare(ar1, ar2 []uint64) int {
	var ln1, ln2, ln int
	ln1 = len(ar1)
	ln2 = len(ar2)
	if ln1 > ln2 {
		ln = ln2
	} else {
		ln = ln1
	}
	for i := 0; i < ln; i++ {
		var cmprl int
		if ar1[i] == ar2[i] {
			cmprl = 0
		} else if ar1[i] < ar2[i] {
			cmprl = -1
		} else if ar1[i] > ar2[i] {
			cmprl = 1
		}
		if cmprl != 0 {
			return cmprl
		}
	}
	if ln1 == ln2 {
		return 0
	} else if ln1 > ln2 {
		return 1
	} else {
		return -1
	}
}

func Uint32ListCompare(ar1, ar2 []uint32) int {
	var ln1, ln2, ln int
	ln1 = len(ar1)
	ln2 = len(ar2)
	if ln1 > ln2 {
		ln = ln2
	} else {
		ln = ln1
	}
	for i := 0; i < ln; i++ {
		var cmprl int
		if ar1[i] == ar2[i] {
			cmprl = 0
		} else if ar1[i] < ar2[i] {
			cmprl = -1
		} else if ar1[i] > ar2[i] {
			cmprl = 1
		}
		if cmprl != 0 {
			return cmprl
		}
	}
	if ln1 == ln2 {
		return 0
	} else if ln1 > ln2 {
		return 1
	} else {
		return -1
	}
}

func Int64ListCompare(ar1, ar2 []int64) int {
	var ln1, ln2, ln int
	ln1 = len(ar1)
	ln2 = len(ar2)
	if ln1 > ln2 {
		ln = ln2
	} else {
		ln = ln1
	}
	for i := 0; i < ln; i++ {
		var cmprl int
		if ar1[i] == ar2[i] {
			cmprl = 0
		} else if ar1[i] < ar2[i] {
			cmprl = -1
		} else if ar1[i] > ar2[i] {
			cmprl = 1
		}
		if cmprl != 0 {
			return cmprl
		}
	}
	if ln1 == ln2 {
		return 0
	} else if ln1 > ln2 {
		return 1
	} else {
		return -1
	}
}

func Int32ListCompare(ar1, ar2 []int32) int {
	var ln1, ln2, ln int
	ln1 = len(ar1)
	ln2 = len(ar2)
	if ln1 > ln2 {
		ln = ln2
	} else {
		ln = ln1
	}
	for i := 0; i < ln; i++ {
		var cmprl int
		if ar1[i] == ar2[i] {
			cmprl = 0
		} else if ar1[i] < ar2[i] {
			cmprl = -1
		} else if ar1[i] > ar2[i] {
			cmprl = 1
		}
		if cmprl != 0 {
			return cmprl
		}
	}
	if ln1 == ln2 {
		return 0
	} else if ln1 > ln2 {
		return 1
	} else {
		return -1
	}
}

func IntListCompare(ar1, ar2 []int) int {
	var ln1, ln2, ln int
	ln1 = len(ar1)
	ln2 = len(ar2)
	if ln1 > ln2 {
		ln = ln2
	} else {
		ln = ln1
	}
	for i := 0; i < ln; i++ {
		var cmprl int
		if ar1[i] == ar2[i] {
			cmprl = 0
		} else if ar1[i] < ar2[i] {
			cmprl = -1
		} else if ar1[i] > ar2[i] {
			cmprl = 1
		}
		if cmprl != 0 {
			return cmprl
		}
	}
	if ln1 == ln2 {
		return 0
	} else if ln1 > ln2 {
		return 1
	} else {
		return -1
	}
}

func String1DCompare(ar1, ar2 []string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if ar1[i] != ar2[i] {
			return false
		}
	}
	return true
}

func String2DCompare(ar1, ar2 [][]string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if String1DCompare(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func String3DCompare(ar1, ar2 [][][]string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if String2DCompare(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func String4DCompare(ar1, ar2 [][][][]string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if String3DCompare(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func StringCompareInsensitive(ar1, ar2 string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if unicode.ToLower(rune(ar1[i])) != unicode.ToLower(rune(ar2[i])) {
			return false
		}
	}
	return true
}

func String1DCompareInsensitive(ar1, ar2 []string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if !StringCompareInsensitive(ar1[i], ar2[i]) {
			return false
		}
	}
	return true
}

func String2DCompareInsensitive(ar1, ar2 [][]string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if String1DCompareInsensitive(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func String3DCompareInsensitive(ar1, ar2 [][][]string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if String2DCompareInsensitive(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func String4DCompareInsensitive(ar1, ar2 [][][][]string) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if String3DCompareInsensitive(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func Byte1DCompare(ar1, ar2 []byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if ar1[i] != ar2[i] {
			return false
		}
	}
	return true
}

func Byte2DCompare(ar1, ar2 [][]byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if Byte1DCompare(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func Byte3DCompare(ar1, ar2 [][][]byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if Byte2DCompare(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func Byte4DCompare(ar1, ar2 [][][][]byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if Byte3DCompare(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func Byte1DCompareInsensitive(ar1, ar2 []byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if unicode.ToLower(rune(ar1[i])) != unicode.ToLower(rune(ar2[i])) {
			return false
		}
	}
	return true
}

func Byte2DCompareInsensitive(ar1, ar2 [][]byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if Byte1DCompareInsensitive(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func Byte3DCompareInsensitive(ar1, ar2 [][][]byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if Byte2DCompareInsensitive(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func Byte4DCompareInsensitive(ar1, ar2 [][][][]byte) bool {
	if len(ar1) != len(ar2) {
		return false
	}
	for i := 0; i < len(ar1); i++ {
		if Byte3DCompareInsensitive(ar1[i], ar2[i]) == false {
			return false
		}
	}
	return true
}

func RemoveAll(set interface{}, val interface{}) interface{} {
	switch set.(type) {
	case [][]byte:
		{
			for i := len(set.([][]byte)) - 1; i >= 0; i-- {
				if Byte1DCompare(set.([][]byte)[i], val.([]byte)) {
					set = append(set.([][]byte)[:i], set.([][]byte)[i+1:]...)
				}
			}
		}
	case [][]string:
		{
			for i := len(set.([][]string)) - 1; i >= 0; i-- {
				if String1DCompare(set.([][]string)[i], val.([]string)) {
					set = append(set.([][]string)[:i], set.([][]string)[i+1:]...)
				}
			}
		}
	case []int:
		{
			for i := len(set.([]int)) - 1; i >= 0; i-- {
				if set.([]int)[i] == val.(int) {
					set = append(set.([]int)[:i], set.([]int)[i+1:]...)
				}
			}
		}
	case []int8:
		{
			for i := len(set.([]int8)) - 1; i >= 0; i-- {
				if set.([]int8)[i] == val.(int8) {
					set = append(set.([]int8)[:i], set.([]int8)[i+1:]...)
				}
			}
		}
	case []int16:
		{
			for i := len(set.([]int16)) - 1; i >= 0; i-- {
				if set.([]int16)[i] == val.(int16) {
					set = append(set.([]int16)[:i], set.([]int16)[i+1:]...)
				}
			}
		}
	case []int32:
		{
			for i := len(set.([]int32)) - 1; i >= 0; i-- {
				if set.([]int32)[i] == val.(int32) {
					set = append(set.([]int32)[:i], set.([]int32)[i+1:]...)
				}
			}
		}
	case []int64:
		{
			for i := len(set.([]int64)) - 1; i >= 0; i-- {
				if set.([]int64)[i] == val.(int64) {
					set = append(set.([]int64)[:i], set.([]int64)[i+1:]...)
				}
			}
		}
	case []uint8:
		{
			for i := len(set.([]uint8)) - 1; i >= 0; i-- {
				if set.([]uint8)[i] == val.(uint8) {
					set = append(set.([]uint8)[:i], set.([]uint8)[i+1:]...)
				}
			}
		}
	case []uint16:
		{
			for i := len(set.([]uint16)) - 1; i >= 0; i-- {
				if set.([]uint16)[i] == val.(uint16) {
					set = append(set.([]uint16)[:i], set.([]uint16)[i+1:]...)
				}
			}

		}
	case []uint32:
		{
			for i := len(set.([]uint32)) - 1; i >= 0; i-- {
				if set.([]uint32)[i] == val.(uint32) {
					set = append(set.([]uint32)[:i], set.([]uint32)[i+1:]...)
				}
			}

		}
	case []uint64:
		{
			for i := len(set.([]uint64)) - 1; i >= 0; i-- {
				if set.([]uint64)[i] == val.(uint64) {
					set = append(set.([]uint64)[:i], set.([]uint64)[i+1:]...)
				}
			}
		}
	case []string:
		{
			for i := len(set.([]string)) - 1; i >= 0; i-- {
				if set.([]string)[i] == val.(string) {
					set = append(set.([]string)[:i], set.([]string)[i+1:]...)
				}
			}
		}
	default:
		panic("error")
	}
	return set
}

func MapCompare(map1 interface{}, map2 interface{}) bool {
	switch map1.(type) {
	case map[int]int:
		{
			for key, val := range map1.(map[int]int) {
				val2, bkey := map2.(map[int]int)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[int]int) {
				val2, bkey := map1.(map[int]int)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[int8]int8:
		{
			for key, val := range map1.(map[int8]int8) {
				val2, bkey := map2.(map[int8]int8)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[int8]int8) {
				val2, bkey := map1.(map[int8]int8)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[int16]int16:
		{
			for key, val := range map1.(map[int16]int16) {
				val2, bkey := map2.(map[int16]int16)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[int16]int16) {
				val2, bkey := map1.(map[int16]int16)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[int32]int32:
		{
			for key, val := range map1.(map[int32]int32) {
				val2, bkey := map2.(map[int32]int32)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[int32]int32) {
				val2, bkey := map1.(map[int32]int32)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[int64]int64:
		{
			for key, val := range map1.(map[int64]int64) {
				val2, bkey := map2.(map[int64]int64)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[int64]int64) {
				val2, bkey := map1.(map[int64]int64)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[uint8]uint8:
		{
			for key, val := range map1.(map[uint8]uint8) {
				val2, bkey := map2.(map[uint8]uint8)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[uint8]uint8) {
				val2, bkey := map1.(map[uint8]uint8)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[uint16]uint16:
		{
			for key, val := range map1.(map[uint16]uint16) {
				val2, bkey := map2.(map[uint16]uint16)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[uint16]uint16) {
				val2, bkey := map1.(map[uint16]uint16)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[uint32]uint32:
		{
			for key, val := range map1.(map[uint32]uint32) {
				val2, bkey := map2.(map[uint32]uint32)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[uint32]uint32) {
				val2, bkey := map1.(map[uint32]uint32)[key]
				if !bkey || val2 != val {
					return false
				}
			}

		}
	case map[uint64]uint64:
		{
			for key, val := range map1.(map[uint64]uint64) {
				val2, bkey := map2.(map[uint64]uint64)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[uint64]uint64) {
				val2, bkey := map1.(map[uint64]uint64)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[string]string:
		{
			for key, val := range map1.(map[string]string) {
				val2, bkey := map2.(map[string]string)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[string]string) {
				val2, bkey := map1.(map[string]string)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[string]int:
		{
			for key, val := range map1.(map[string]int) {
				val2, bkey := map2.(map[string]int)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[string]int) {
				val2, bkey := map1.(map[string]int)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	case map[int]string:
		{
			for key, val := range map1.(map[int]string) {
				val2, bkey := map2.(map[int]string)[key]
				if !bkey || val2 != val {
					return false
				}
			}
			for key, val := range map2.(map[int]string) {
				val2, bkey := map1.(map[int]string)[key]
				if !bkey || val2 != val {
					return false
				}
			}
		}
	default:
		panic("error")
	}
	return true
}

func SeperateUseList(srcstr string, sepwith []string) []string {
	sepresult := []string{}
	srcstrbt := []byte(srcstr)
	temp := []byte{}
	for i := 0; i < len(srcstrbt); {
		fndcnt := 0
		for _, sep := range sepwith {
			if string(srcstrbt[i:i+len(sep)]) == string(sep) {
				if len(temp) > 0 {
					sepresult = append(sepresult, string(temp))
					temp = []byte{}
				}
				sepresult = append(sepresult, sep)
				fndcnt = len([]byte(sep))
			}
		}
		if fndcnt == 0 {
			temp = append(temp, srcstrbt[i])
			i++
		} else {
			i += fndcnt
		}
	}
	if len(temp) > 0 {
		sepresult = append(sepresult, string(temp))
		temp = []byte{}
	}
	return sepresult
}

func SkipWhiteSpace(bt []byte, i int) int {
	if i >= len(bt) {
		return i
	}
	for bt[i] == '\r' || bt[i] == '\n' || bt[i] == '\t' || bt[i] == ' ' {
		i++
	}
	return i
}

func Until(bt []byte, i int, cbt []byte) int {
	if i >= len(bt) {
		return i
	}
	for true {
		if bytes.Index(cbt, []byte{bt[i]}) != -1 {
			break
		}
		i++
	}
	return i
}

//scriptobj = line + line seperate
func BoscriptSeperate(script string) (scriptobj []string) {
	script = strings.Trim(script, "\r\n\t ")
	if script[0] != '{' && script[len(script)-1] != '}' {
		script = "{" + script + "}"
	}
	scriptbt := []byte(script)
	temp := []byte{}
	bracket_stack := []string{}
	bracket_start := []int{}
	bracket_cur := []int{}
	bracket_waitfor := []string{}
	bracket_in := []string{}
	i := 0
	for i < len(scriptbt) {
		gospecialnext := false
		i = SkipWhiteSpace(scriptbt, i)
		if i < len(scriptbt) && (scriptbt[i] == '{' || bracket_stack[len(bracket_stack)-1] == "{") {
			gospecialnext = false
			binherecnt := 0
			for scriptbt[i] == '{' || bracket_stack[len(bracket_stack)-1] == "{" {
				gonextstack := false
				for scriptbt[i] != '}' {
					switch scriptbt[i] {
					case '{':
						if binherecnt > 0 {
							bracket_waitfor[len(bracket_waitfor)-1] = "}"
						}

						bracket_stack = append(bracket_stack, "{")
						bracket_start = append(bracket_start, i)
						bracket_cur = append(bracket_cur, i)
						bracket_waitfor = append(bracket_waitfor, "")
						bracket_in = append(bracket_in, "")
						if binherecnt > 0 {
							gonextstack = true
						} else {
							i++
						}
						binherecnt += 1
						break
					case '[':
						bracket_waitfor[len(bracket_waitfor)-1] = "]"

						gospecialnext = true
						break
					case '(':
						bracket_waitfor[len(bracket_waitfor)-1] = ")"

						gospecialnext = true
						break
					case '.':
						temp = bytes.Trim(temp, "\r\n\t ")
						if len(temp) > 0 {
							for true {
								i = Until(scriptbt, i, []byte{'.', '='})
								if scriptbt[i] == '=' {
									break
								} else {
									i += 1
								}
							}
							i = SkipWhiteSpace(scriptbt, i)
						}
					case '=':
						temp = bytes.Trim(temp, "\r\n\t ")
						if len(temp) > 0 {
							scriptobj = append(scriptobj, string(temp))

							bracket_waitfor[len(bracket_waitfor)-1] = "wait set line end"
							i += 1
							temp = []byte{}
							i = SkipWhiteSpace(scriptbt, i)
						}
					default:
						temp = append(temp, scriptbt[i])
						i++
					}
					if gonextstack {
						break
					}
					if gospecialnext {
						break
					}
				}
				if gonextstack {
					continue
				} else if gospecialnext {
					break
				} else {
					for true {
						i += 1
						i = SkipWhiteSpace(scriptbt, i)

						bracket_stack = bracket_stack[:len(bracket_stack)-1]
						bracket_start = bracket_start[:len(bracket_start)-1]
						bracket_cur = bracket_cur[:len(bracket_cur)-1]
						bracket_waitfor = bracket_waitfor[:len(bracket_waitfor)-1]
						bracket_in = bracket_in[:len(bracket_in)-1]

						if i < len(scriptbt) && string(bracket_stack[len(bracket_stack)-1]) == "{" && string([]byte{scriptbt[i]}) == bracket_waitfor[len(bracket_waitfor)-1] {
							continue
						} else {
							break
						}
					}
					if i < len(scriptbt) && scriptbt[i] == ',' {
						i += 1
						i = SkipWhiteSpace(scriptbt, i)
					}
					break
				}
			}
		}
		if i < len(scriptbt) && (scriptbt[i] == '[' || bracket_stack[len(bracket_stack)-1] == "[") {
			gospecialnext = false
			binherecnt := 0
			for scriptbt[i] == '[' || bracket_stack[len(bracket_stack)-1] == "[" {
				i = SkipWhiteSpace(scriptbt, i)
				gonextstack := false
				for scriptbt[i] != ']' {
					switch scriptbt[i] {
					case '{':
						bracket_waitfor[len(bracket_waitfor)-1] = "}"

						gospecialnext = true
						break
					case '[':
						if binherecnt > 0 {
							bracket_waitfor[len(bracket_waitfor)-1] = "]"
						}

						bracket_stack = append(bracket_stack, "[")
						bracket_start = append(bracket_start, i)
						bracket_cur = append(bracket_cur, i)
						bracket_waitfor = append(bracket_waitfor, "")
						bracket_in = append(bracket_in, "")

						if binherecnt > 0 {
							gonextstack = true
						} else {
							i++
						}
						binherecnt += 1
						break
					case '(':
						bracket_waitfor[len(bracket_waitfor)-1] = ")"

						gospecialnext = true
						break
					case '.':
						temp = bytes.Trim(temp, "\r\n\t ")
						if len(temp) > 0 {
							for true {
								i = Until(scriptbt, i, []byte{'.', '='})
								if scriptbt[i] == '=' {
									break
								} else {
									i += 1
								}
							}
							i = SkipWhiteSpace(scriptbt, i)
						}
					default:
						temp = append(temp, scriptbt[i])
						i++
					}
					if gonextstack {
						break
					}
					if gospecialnext {
						break
					}
				}
				if gonextstack {
					continue
				} else if gospecialnext {
					break
				} else {
					for true {
						i += 1
						i = SkipWhiteSpace(scriptbt, i)

						bracket_stack = bracket_stack[:len(bracket_stack)-1]
						bracket_start = bracket_start[:len(bracket_start)-1]
						bracket_cur = bracket_cur[:len(bracket_cur)-1]
						bracket_waitfor = bracket_waitfor[:len(bracket_waitfor)-1]
						bracket_in = bracket_in[:len(bracket_in)-1]

						if i < len(scriptbt) && string(bracket_stack[len(bracket_stack)-1]) == "[" && string([]byte{scriptbt[i]}) == bracket_waitfor[len(bracket_waitfor)-1] {
							continue
						} else {
							break
						}
					}
					break
				}
			}
		}
		if i < len(scriptbt) && (scriptbt[i] == '(' || bracket_stack[len(bracket_stack)-1] == "(") {
			gospecialnext = false
			binherecnt := 0
			for scriptbt[i] == '(' || bracket_stack[len(bracket_stack)-1] == "(" {
				i = SkipWhiteSpace(scriptbt, i)
				gonextstack := false
				for !(scriptbt[i] == ')' || scriptbt[i] == '}') {
					switch scriptbt[i] {
					case '{':
						bracket_waitfor[len(bracket_waitfor)-1] = "}"

						gospecialnext = true
						break
					case '[':
						bracket_waitfor[len(bracket_waitfor)-1] = "]"

						gospecialnext = true
						break
					case '(':
						if binherecnt > 0 {
							bracket_waitfor[len(bracket_waitfor)-1] = ")"
						}

						bracket_stack = append(bracket_stack, "(")
						bracket_start = append(bracket_start, i)
						bracket_cur = append(bracket_cur, i)
						bracket_waitfor = append(bracket_waitfor, ")")
						bracket_in = append(bracket_in, "")

						if binherecnt > 0 {
							gonextstack = true
						} else {
							i++
						}
						binherecnt += 1
						break
					case '.':
						temp = bytes.Trim(temp, "\r\n\t ")
						if len(temp) > 0 {
							for true {
								i = Until(scriptbt, i, []byte{'.', '='})
								if scriptbt[i] == '=' {
									break
								} else {
									i += 1
								}
							}
							i = SkipWhiteSpace(scriptbt, i)
						}
					case '=':
						temp = bytes.Trim(temp, "\r\n\t ")
						if len(temp) > 0 {
							bracket_waitfor[len(bracket_waitfor)-1] = "wait set line end"
							i += 1
							temp = []byte{}
							i = SkipWhiteSpace(scriptbt, i)
						}
					default:
						temp = append(temp, scriptbt[i])
						i++
					}
					if gonextstack {
						break
					}
					if gospecialnext {
						break
					}
				}
				if gonextstack {
					continue
				} else if gospecialnext {
					break
				} else {
					for true {
						i += 1
						i = SkipWhiteSpace(scriptbt, i)

						bracket_stack = bracket_stack[:len(bracket_stack)-1]
						bracket_start = bracket_start[:len(bracket_start)-1]
						bracket_cur = bracket_cur[:len(bracket_cur)-1]
						bracket_waitfor = bracket_waitfor[:len(bracket_waitfor)-1]
						bracket_in = bracket_in[:len(bracket_in)-1]

						if i < len(scriptbt) && string(bracket_stack[len(bracket_stack)-1]) == "(" && string([]byte{scriptbt[i]}) == bracket_waitfor[len(bracket_waitfor)-1] {
							continue
						} else {
							break
						}
					}
					break
				}
			}
		}
	}
	return scriptobj
}

func GlobalIsoDateSecondTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GlobalIsoDateMilisecTime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func GlobalIsoDateMicrosecTime() string {
	return time.Now().Format("2006-01-02 15:04:05.000000")
}

func GlobalIsoDateNanosecTime() string {
	return time.Now().Format("2006-01-02 15:04:05.000000000")
}

func GlobalIsoDay() string {
	return time.Now().Format("2006-01-02")
}

func GlobalIsoSecondTime() string {
	return time.Now().Format("15:04:05.000000")
}

func GlobalIsoMilisecTime() string {
	return time.Now().Format("15:04:05.000")
}

func GlobalIsoMicrosecTime() string {
	return time.Now().Format("15:04:05.000000")
}

func GlobalIsoNanosecTime() string {
	return time.Now().Format("15:04:05.000000000")
}

func CurTime() string {
	return time.Now().Local().Format("2006-01-02 15:04:05")
}

func LocalTime() string {
	return time.Now().Local().Format("2006-01-02 15:04:05")
}

func TodayDateTimeShortStr() string {
	return time.Now().Local().Format("20060102150405")
}

func TodayDateShortStr() string {
	return time.Now().Local().Format("20060102")
}

func TodayTimeShortStr() string {
	return time.Now().Local().Format("150405")
}

func TimeForFileName() string {
	return strings.Replace(isotime.GlobalIsoDateMicrosecTime(), ":", "_", -1)
}

func IsoDateTime() string {
	return time.Now().Local().Format("2006-01-02T15:04:05")
}

func TodayISOTime() string {
	return time.Now().Local().Format("2006-01-02T15:04:05")
}
func TodayIsoDate() string {
	return time.Now().Local().Format("2006-01-02")
}

func TodayDate() string {
	return time.Now().Local().Format("2006-01-02")
}

func ISOTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}
func IsoDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func Date(t time.Time) string {
	return t.Format("2006-01-02")
}

func Time() string {
	return time.Now().Local().Format("15:04:05")
}

func WorldTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func LocalTimeSecond() int64 {
	return time.Now().Local().Unix()
}

func WorldTimeSecond() int64 {
	return time.Now().Unix()
}

func LocalTimeSecondStr() string {
	return strconv.FormatInt(time.Now().Local().Unix(), 10)
}

func WorldTimeSecondStr() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func AbbreviateMonthToNum(monshortword string) int {
	var mont int
	switch monshortword {
	case "Jan":
		mont = 1
	case "Feb":
		mont = 2
	case "Mar":
		mont = 3
	case "Apr":
		mont = 4
	case "May":
		mont = 5
	case "June":
		mont = 6
	case "July":
		mont = 7
	case "Aug":
		mont = 8
	case "Sept":
		mont = 9
	case "Oct":
		mont = 10
	case "Nov":
		mont = 11
	case "Dec":
		mont = 12
	default:
		mont = -1
	}
	return mont
}

func GetMatch(matchpage []byte, mainds []int, matchwith string) []byte {
	matchwithbt := []byte(matchwith)
	matchresult := []byte{}
	for mwi := 0; mwi < len(matchwithbt); {
		if mwi+1 < len(matchwithbt) && unicode.IsDigit(rune(matchwithbt[mwi+1])) && matchwithbt[mwi] == '$' {
			indpos, _ := strconv.ParseInt(string([]byte{matchwithbt[mwi+1]}), 10, 64)
			matchresult = append(matchresult, matchpage[mainds[2*indpos]:mainds[2*indpos+1]]...)
			mwi += 2
		} else {
			matchresult = append(matchresult, matchwithbt[mwi])
			mwi++
		}
	}
	return matchresult
}

func GetMatchWithSlashOp(matchpage []byte, mainds []int, matchwith string) []byte {
	matchwithbt := []byte(matchwith)
	matchresult := []byte{}
	for mwi := 0; mwi < len(matchwithbt); {
		if mwi+1 < len(matchwithbt) && unicode.IsDigit(rune(matchwithbt[mwi+1])) && matchwithbt[mwi] == '\\' {
			indpos, _ := strconv.ParseInt(string([]byte{matchwithbt[mwi+1]}), 10, 64)
			matchresult = append(matchresult, matchpage[mainds[2*indpos]:mainds[2*indpos+1]]...)
			mwi += 2
		} else {
			matchresult = append(matchresult, matchwithbt[mwi])
			mwi++
		}
	}
	return matchresult
}

func HttpParam(r *http.Request, name string) string {
	value2 := r.PostFormValue(name)
	if len(value2) > 0 {
		return value2
	}
	result := r.URL.Query().Get(name)
	return result
}

func UnitTimeToSecond(unittime string) uint64 {
	RefreshInterval2 := uint64(0)
	if strings.HasSuffix(string(unittime), "day") {
		RefreshInterval2, _ = strconv.ParseUint(string(unittime[:len(unittime)-3]), 10, 64)
		RefreshInterval2 *= 24 * 3600
	} else if strings.HasSuffix(string(unittime), "hour") {
		RefreshInterval2, _ = strconv.ParseUint(string(unittime[:len(unittime)-4]), 10, 64)
		RefreshInterval2 *= 3600
	} else if strings.HasSuffix(string(unittime), "minute") {
		RefreshInterval2, _ = strconv.ParseUint(string(unittime[:len(unittime)-6]), 10, 64)
		RefreshInterval2 *= 60
	} else if strings.HasSuffix(string(unittime), "second") {
		RefreshInterval2, _ = strconv.ParseUint(string(unittime[:len(unittime)-6]), 10, 64)
	} else {
		RefreshInterval2, _ = strconv.ParseUint(string(unittime[:]), 10, 64)
	}
	return RefreshInterval2
}

func Float64Clone(val []float64) []float64 {
	val2 := make([]float64, len(val))
	copy(val2, val)
	return val2
}

//rng must like {0.1,0.2,0.2,0.3,0.9,1}
func RangeIndex(rng, value interface{}) (rangeindex int) {
	rangeindex = -1
	switch rng.(type) {
	case []int:
		if len(rng.([]int))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]int))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]int)[2*cur]
			valend := rng.([]int)[2*cur+1]
			if value.(int) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []int8:
		if len(rng.([]int8))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]int8))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]int8)[2*cur]
			valend := rng.([]int8)[2*cur+1]
			if value.(int8) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int8) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []int16:
		if len(rng.([]int16))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]int16))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]int16)[2*cur]
			valend := rng.([]int16)[2*cur+1]
			if value.(int16) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int16) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []int32:
		if len(rng.([]int32))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]int32))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]int32)[2*cur]
			valend := rng.([]int32)[2*cur+1]
			if value.(int32) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int32) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []int64:
		if len(rng.([]int64))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]int64))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]int64)[2*cur]
			valend := rng.([]int64)[2*cur+1]
			if value.(int64) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int64) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []uint8:
		if len(rng.([]uint8))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]uint8))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]uint8)[2*cur]
			valend := rng.([]uint8)[2*cur+1]
			if value.(uint8) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint8) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []uint16:
		if len(rng.([]uint16))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]uint16))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]uint16)[2*cur]
			valend := rng.([]uint16)[2*cur+1]
			if value.(uint16) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint16) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []uint32:
		if len(rng.([]uint32))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]uint32))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]uint32)[2*cur]
			valend := rng.([]uint32)[2*cur+1]
			if value.(uint32) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint32) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []uint64:
		if len(rng.([]uint64))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]uint64))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]uint64)[2*cur]
			valend := rng.([]uint64)[2*cur+1]
			if value.(uint64) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint64) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []float32:
		if len(rng.([]float32))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]float32))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]float32)[2*cur]
			valend := rng.([]float32)[2*cur+1]
			if value.(float32) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(float32) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []float64:
		if len(rng.([]float64))%2 != 0 {
			return -1
		}
		start := 0
		end := len(rng.([]float64))/2 - 1
		cur := end / 2
		for true {
			valstart := rng.([]float64)[2*cur]
			valend := rng.([]float64)[2*cur+1]
			if value.(float64) < valstart {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(float64) >= valend {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	default:
		panic("unsupport error")
	}
	return -1
}

func HttpAttachmentName(head http.Header) string {
	for key, val := range head {
		if key == "Content-Disposition" {
			if strings.HasPrefix(val[0], "attachment; filename*=") {
				codename := val[0][len("attachment; filename*="):strings.Index(val[0], "''")]
				resUri, pErr := url.Parse(val[0][strings.Index(val[0], "''")+2:])
				if pErr != nil {
					return ""
				}
				dc := mahonia.NewDecoder(codename)
				return dc.ConvertString(resUri.RawQuery)
			} else if strings.HasPrefix(val[0], "attachment; filename=") {
				resUri, pErr := url.Parse(val[0][len("attachment; filename="):])
				if pErr != nil {
					return ""
				}
				return string(resUri.Path)
			}
		}
	}
	return ""
}

func HttpContentLength(head http.Header) int64 {
	for key, val := range head {
		if key == "Content-Length" {
			contentlen, bcontentlen := strconv.ParseInt(val[0], 10, 64)
			if bcontentlen == nil {
				return contentlen
			}
		}
	}
	return -1
}

func GetRegexGroup1(regex, str string) string {
	re, reerr := regexp.Compile(regex)
	if reerr == nil {
		inds := re.FindAllSubmatchIndex([]byte(str), -1)
		if len(inds) > 0 && len(inds[0]) >= 4 {
			return str[inds[0][2]:inds[0][3]]
		}
	}
	return ""
}

func IsHomeUrl(url string) bool {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		endslashpos := strings.Index(url[9:], "/")
		if endslashpos != -1 {
			if strings.LastIndex(url, "/") == 9+endslashpos && url[len(url)-1] == '/' {
				return true
			}
		} else {
			return true
		}
	}
	return false
}
func StdHomeUrl(url string) string {
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return ""
	}
	domainend := strings.Index(url[9:], "/")
	if domainend != -1 {
		return strings.ToLower(url[:9+domainend+1])
	} else {
		return strings.ToLower(url + "/")
	}
}
func GetHomeUrl(url string) string {
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return ""
	}
	domainend := strings.Index(url[9:], "/")
	if domainend != -1 {
		return strings.ToLower(url[:9+domainend+1])
	} else {
		return strings.ToLower(url + "/")
	}
}

func StdUrlLinkDomain(url string) string {
	if strings.EqualFold(url, "http://") || strings.EqualFold(url, "https://") {
		lastslashi := strings.Index(url[7:], "/")
		if lastslashi != -1 {
			return strings.ToLower(url[:7+lastslashi]) + url[7+lastslashi:]
		} else {
			return strings.ToLower(url) + "/"
		}
	}
	return url
}

func GetNoTagHomeUrl(url string) (notaghomneurl string) {
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return ""
	}
	if strings.Index(url[8:], "/") == -1 {
		if strings.LastIndex(url, "/") != -1 {
			notaghomneurl = url[strings.LastIndex(url, "/")+1:]
			if strings.Index(notaghomneurl, ":") != -1 {
				notaghomneurl = notaghomneurl[:strings.Index(notaghomneurl, ":")]
			}
			return notaghomneurl
		} else {
			return ""
		}
	}
	domainend := strings.Index(url[8:], "/")
	if domainend != -1 {
		url = url[:8+domainend]
		notaghomneurl = strings.ToLower(url[strings.Index(url, "://")+3:])
		if strings.Index(notaghomneurl, ":") != -1 {
			notaghomneurl = notaghomneurl[:strings.Index(notaghomneurl, ":")]
		}
		return notaghomneurl
	} else {
		notaghomneurl = strings.ToLower(url[strings.Index(url, "://")+3:])
		if strings.Index(notaghomneurl, ":") != -1 {
			notaghomneurl = notaghomneurl[:strings.Index(notaghomneurl, ":")]
		}
		return notaghomneurl
	}
}

func DomainMatch(url1, url2 string) bool {
	if !(strings.HasPrefix(url1, "http://") || strings.HasPrefix(url1, "https://")) || !(strings.HasPrefix(url2, "http://") || strings.HasPrefix(url2, "https://")) {
		return false
	}
	if strings.Index(url1[8:], "/") == -1 {
		url1 += "/"
	}
	if strings.Index(url2[8:], "/") == -1 {
		url2 += "/"
	}
	url1h := GetNoTagHomeUrl(url1)
	url2h := GetNoTagHomeUrl(url2)
	i2 := len(url2h) - 1
	dotcnt := 0
	i := len(url1h) - 1
	for i >= 0 && i2 >= 0 {
		if url1h[i] == url2h[i2] {
			if url1h[i] == '.' {

				dotcnt++
			}
			if i == 0 && i2 == 0 {
				return true
			}
			i--
			i2--
		} else {
			break
		}
	}
	if dotcnt >= 3 {
		return true
	} else if dotcnt == 2 {
		url1h = url1h[:strings.LastIndex(url1h, ".")]
		url2h = url2h[:strings.LastIndex(url2h, ".")]
		topdomain := []string{".com", ".xyz", ".net", ".top", ".tech", ".org", ".gov", ".edu", ".ink", ".red", ".int", ".mil", ".pub", ".cc"}
		for _, topdomain := range topdomain {
			if strings.HasSuffix(url1h, topdomain) {
				return false
			}
		}
		return true
	}
	return false
}

func GetUrlProtocol(url1 string) string {
	if strings.HasPrefix(url1, "http://") {
		return "http://"
	}
	if strings.HasPrefix(url1, "https://") {
		return "https://"
	}
	return ""
}

func GetUrlDomain(url1 string) string {
	if !(strings.HasPrefix(url1, "http://") || strings.HasPrefix(url1, "https://")) {
		return ""
	}
	if strings.Index(url1[8:], "/") == -1 {
		url1 += "/"
	} else {
		url1 = url1[:8+strings.Index(url1[8:], "/")+1]
	}
	url1h := GetNoTagHomeUrl(url1)
	if strings.LastIndex(url1h, ".") == -1 {
		return ""
	}
	domain := url1h[strings.LastIndex(url1h, "."):]
	url1h = url1h[:strings.LastIndex(url1h, ".")]
	topdomain := []string{".com", ".xyz", ".net", ".top", ".tech", ".org", ".gov", ".edu", ".ink", ".red", ".int", ".mil", ".pub", ".cc"}
	for _, topdomain := range topdomain {
		if strings.HasSuffix(url1h, topdomain) {
			domain = url1h[strings.LastIndex(url1h, "."):] + domain
			url1h = url1h[:strings.LastIndex(url1h, ".")]
			break
		}
	}
	if strings.LastIndex(url1h, ".") != -1 {
		domain = strings.Trim(url1h[strings.LastIndex(url1h, ".")+1:]+domain, "\r\n\t ")
		if strings.Index(domain, "?") != -1 {
			return ""
		}
		if strings.Index(domain, ":") != -1 {
			return domain[:strings.Index(domain, ":")]
		}
		return domain
	} else {
		retdomain := strings.Trim(url1h+domain, "\r\b\t ")
		if strings.Index(retdomain, "?") != -1 {
			return ""
		}
		if strings.Index(domain, ":") != -1 {
			return domain[:strings.Index(domain, ":")]
		}
		return retdomain
	}
}

func CreateDir(path string, perm os.FileMode) bool {
	dirnamedeque := regexp.MustCompile("[/\\\\]+").Split(path, -1)
	path = ""
	for i := 0; i < len(dirnamedeque); i++ {
		path += dirnamedeque[i] + "/"
		os.Mkdir(path, perm)
	}
	return true
}

func PathSplit(path string) []string {
	pathitems := regexp.MustCompile("[/\\\\]+").Split(path, -1)
	if len(pathitems) > 0 && pathitems[0] == "" {
		pathitems = pathitems[1:]
	}
	if len(pathitems) > 0 && pathitems[len(pathitems)-1] == "" {
		pathitems = pathitems[0 : len(pathitems)-1]
	}
	return pathitems
}

func MoveFile(path1, path2 string) bool {
	err := os.Rename(path1, path2)
	if err == nil {
		return true
	}
	ff, ffe := os.OpenFile(path2, os.O_WRONLY|os.O_CREATE, 00666)
	if ffe == nil {
		ff2, ff2e := os.OpenFile(path1, os.O_RDONLY, 00666)
		buf := make([]byte, 1048576)
		if ff2e == nil {
			for true {
				rcnd, rcnde := ff2.Read(buf)
				if rcnde != nil {
					break
				}
				ff.Write(buf[:rcnd])
			}
			ff2.Close()
		} else {
			ff.Close()
			return false
		}
		ff.Close()
		os.Remove(path1)
		return true
	}
	return false
}

func MoveDir(dir1, dir2 string) bool {
	if !(dir1[len(dir1)-1] == '/' || dir1[len(dir1)-1] == '\\') {
		dir1 += "/"
	}
	if !(dir2[len(dir2)-1] == '/' || dir2[len(dir2)-1] == '\\') {
		dir2 += "/"
	}
	dir1info, err := ioutil.ReadDir(dir1)
	if err == nil {
		for _, dir1i := range dir1info {
			if dir1i.IsDir() {
				os.Mkdir(dir2+dir1i.Name(), 0666)
				MoveDir(dir1+dir1i.Name(), dir2+dir1i.Name())
			} else {
				os.Rename(dir1+dir1i.Name(), dir2+dir1i.Name())
			}
		}
		os.Remove(dir1)
	} else {
		return false
	}
	return true
}

func AppDir() string {
	appdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	appdir = strings.Replace(appdir, "\\", "/", -1)
	if appdir[len(appdir)-1] != '/' {
		appdir += "/"
	}
	return appdir
}

func IsChineseOrEnglish(ctt []byte) bool {
	return regexp.MustCompile("(?ism)^[a-zA-Z0-9 \\p{Han}\\p{P}]+$").Match([]byte(ctt))
}

func IsEnglish(ctt []byte) bool {
	return regexp.MustCompile("(?ism)^[ -~\r\n]+$").Match([]byte(ctt))
}

func AppParentDir() string {
	appdir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	appdir = strings.Replace(appdir, "\\", "/", -1)
	if appdir[len(appdir)-1] == '/' {
		appdir = appdir[:len(appdir)-1]
	}
	if strings.LastIndex(appdir, "/") != -1 {
		return appdir[:strings.LastIndex(appdir, "/")+1]
	} else {
		return ""
	}
}
func CurDir() string {
	appdir, _ := os.Getwd()
	appdir = strings.Replace(appdir, "\\", "/", -1)
	if appdir[len(appdir)-1] != '/' {
		appdir += "/"
	}
	return appdir
}

func CurParentDir() string {
	appdir, _ := os.Getwd()
	appdir = strings.Replace(appdir, "\\", "/", -1)
	if appdir[len(appdir)-1] == '/' {
		appdir = appdir[:len(appdir)-1]
	}
	appdir = appdir[:strings.LastIndex(appdir, "/")+1]
	appdir += "/"
	return appdir
}

func StdPath(path string) string {
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	return path
}

func StdUnixLikePath(path string) string {
	re := regexp.MustCompile("[/\\\\]+")
	return re.ReplaceAllString(path, "/")
}

func StdDir(path string) string {
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

//without dot and space prefix or suffix.and remove middle <>?/\|*?":
func StdFileName(name string) string {
	return strings.Trim(string(regexp.MustCompile("[\\r\\t\\n:\\?\\\\\\/\\\"<>\\|\\*]+").ReplaceAll([]byte(name), []byte{})), ". ")
}

func ToAbsolutePath(path string) string {
	if path == "" {
		return ""
	}
	var bsepend bool
	if path[len(path)-1] == '/' || path[len(path)-1] == '\\' {
		bsepend = true
	}
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\\\", "\\", -1)
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	curdir := ""
	if runtime.GOOS == "windows" {
		if path[1] != ':' {
			curdir = CurDir()
		}
	} else if path[0] != '/' {
		curdir = CurDir()
	}
	curdirls := strings.Split(curdir, "/")
	if len(curdirls) > 0 {
		curdirls = curdirls[:len(curdirls)-1]
	}
	if path[0] == '/' || path[1] == ':' {
		if len(curdirls) > 0 {
			curdirls = curdirls[:1]
		}
	}

	pathls := strings.Split(path, "/")
	for i := 0; i < len(pathls); i++ {
		if pathls[i] == "." || pathls[i] == "" {

		} else if pathls[i] == ".." {
			if len(pathls)-1 >= 0 {
				curdirls = curdirls[:len(curdirls)-1]
			}
		} else {
			curdirls = append(curdirls, pathls[i])
		}
	}
	okpath := strings.Join(curdirls, "/")
	if bsepend {
		okpath += "/"
	}
	return okpath
}

func UrlToRegex(url string) string {
	var ind int
	if strings.Index(url[8:], "/") != -1 {
		ind = 8 + strings.Index(url[8:], "/")
	}
	for i := 0; i < len(url); {
		if url[i] == '.' || url[i] == '[' || url[i] == ']' || url[i] == '{' || url[i] == '}' || url[i] == '(' || url[i] == ')' || url[i] == '*' || url[i] == '|' {
			url = url[:i] + "\\" + url[i:]
			i += 2
			continue
		}
		if url[i] == '?' {
			for i < len(url) {
				if url[i] != '=' {
					i++
					continue
				} else {
					i += 1
					starti := i
					numcnt := 0
					for i < len(url) {
						if url[i] == '&' {
							break
						}
						if url[i] >= '0' && url[i] <= '9' {
							numcnt += 1
						}
						i++
					}
					if numcnt > 4 && i-starti == numcnt {
						url = url[:starti] + "\\d+" + url[i:]
						i = starti + 3
					} else {
						url = url[:starti] + "[ -~]+" + url[i:]
						i = starti + 6
					}
					i++
				}
			}
		}
		if i >= ind && i < len(url) {
			if url[i] == '%' {
				starti := i
				ll := 0
				for i+2 < len(url) && url[i] == '%' && (url[i+1] >= '0' && url[i+1] <= '9' || url[i+1] >= 'a' && url[i+1] <= 'z' || url[i+1] >= 'A' && url[i+1] <= 'Z') && (url[i+2] >= '0' && url[i+2] <= '9' || url[i+2] >= 'a' && url[i+2] <= 'z' || url[i+2] >= 'A' && url[i+2] <= 'Z') {
					i += 3
					ll += 3
				}
				url = url[:starti] + "[ -~]+" + url[i:]
				i = starti + 6
				continue
			}
			if url[i] >= '0' && url[i] <= '9' {
				starti := i
				ll := 0
				for i < len(url) && url[i] >= '0' && url[i] <= '9' {
					i += 1
					ll += 1
				}
				url = url[:starti] + "\\d+" + url[i:]
				i = starti + 3
				continue
			}
		}
		i++
	}
	if strings.HasSuffix(url, "\\\\") == false && strings.HasSuffix(url, "\\") == true {
		url = url[:len(url)-1]
	}
	return "(?ism)" + url
}

func GetLocalIPs() (localips []string) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localips = append(localips, ipnet.IP.String())
			}
		}
	}
	return localips
}

//通过dns服务器8.8.8.8:80获取使用的ip,maybe have error by router diffrent.
func GetPulicIPByDns() string {
	conn, _ := net.Dial("udp", "114.114.114.114:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}

func GetPulicIPByIp138() string {
	ctt, _, _, code, _ := netutil.UrlGet("http://2019.ip138.com/ic.asp", nil, false, nil, nil, 30*time.Second, 30*time.Second, nil)
	if code == 200 {
		ipctt := regexp.MustCompile("\\d+\\.\\d+\\.\\d+\\.\\d+").Find(ctt)
		return string(ipctt)
	}
	return ""
}

//判断是否是公网ip
func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func CheckZero(bt []byte) bool {
	for i := 0; i < len(bt); i++ {
		if bt[i] != 0 {
			return false
		}
	}
	return true
}

func Packu64u32(u64 uint64, u32 uint32) (u64bt []byte) {
	u64bt = make([]byte, 12)
	binary.BigEndian.PutUint64(u64bt[:8], u64)
	binary.BigEndian.PutUint32(u64bt[8:8+4], u32)
	return u64bt
}

func PackUint64(u64 uint64) (u64bt []byte) {
	u64bt = make([]byte, 8)
	binary.BigEndian.PutUint64(u64bt, u64)
	return u64bt
}

func UnpackUint64(u64bt []byte) uint64 {
	return binary.BigEndian.Uint64(u64bt)
}

func PackUint32(u32 uint32) (u32bt []byte) {
	u32bt = make([]byte, 4)
	binary.BigEndian.PutUint32(u32bt, u32)
	return u32bt
}

func UnpackUint32(u32bt []byte) uint32 {
	return binary.BigEndian.Uint32(u32bt)
}

func PackUint16(u16 uint16) (u16bt []byte) {
	u16bt = make([]byte, 2)
	binary.BigEndian.PutUint16(u16bt, u16)
	return u16bt
}

func UnpackUint16(u16bt []byte) uint16 {
	return binary.BigEndian.Uint16(u16bt)
}

func PackFloat32(f32 float32) (bt []byte) {
	bt = make([]byte, 4)
	binary.BigEndian.PutUint32(bt, math.Float32bits(f32))
	return bt
}

func UnpackFloat32(bt []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(bt))
}

func PackFloat64(f64 float64) (bt []byte) {
	bt = make([]byte, 8)
	binary.BigEndian.PutUint64(bt, math.Float64bits(f64))
	return bt
}

func UnpackFloat64(bt []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(bt))
}

func EndOfLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else if runtime.GOOS == "linux" || runtime.GOOS == "freebsd" {
		return "\n"
	} else if runtime.GOOS == "macos" {
		return "\r"
	} else {
		panic("unkno wsystem")
	}
}

func GetEndOfLine(ctt []byte) string {
	if bytes.Index(ctt, []byte{'\r', '\n'}) != -1 {
		return "\r\n"
	} else if bytes.Index(ctt, []byte{'\n'}) != -1 {
		return "\n"
	} else if bytes.Index(ctt, []byte{'\r'}) != -1 {
		return "\r"
	} else {
		return EndOfLine()
	}
}

func PackFloat32Str(f32str string) (bt []byte) {
	f32, f32e := strconv.ParseFloat(f32str, 32)
	if f32e != nil {
		fmt.Println("PackFloat32Str", f32str)
		panic("error")
	}
	bt = make([]byte, 4)
	binary.BigEndian.PutUint32(bt, math.Float32bits(float32(f32)))
	return bt
}

func PackFloat64Str(f64str string) (bt []byte) {
	f64, f64e := strconv.ParseFloat(f64str, 64)
	if f64e != nil {
		fmt.Println("PackFloat64Str", f64str)
		panic("error")
	}
	bt = make([]byte, 8)
	binary.BigEndian.PutUint64(bt, math.Float64bits(f64))
	return bt
}

func ByteUint16ToNumString(u16bt []byte) string {
	return strconv.FormatInt(int64(binary.BigEndian.Uint16(u16bt)), 10)
}

func ByteUint32ToNumString(u32bt []byte) string {
	return strconv.FormatInt(int64(binary.BigEndian.Uint32(u32bt)), 10)
}

func ByteUint64ToNumString(u64bt []byte) string {
	return strconv.FormatInt(int64(binary.BigEndian.Uint64(u64bt)), 10)
}

func FileNameUnstd(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '%' {
			val, vale := strconv.ParseInt(name[i : i+3][1:], 16, 32)
			if vale == nil {
				name = name[:i] + string([]byte{byte(val)}) + name[i+3:]
			}
		}
	}
	return name
}

func FileNameStd(namedata string) string {
	for i := len(namedata) - 1; i >= 0; i-- {
		if namedata[i] >= 0 && namedata[i] <= 32 || namedata[i] == '<' || namedata[i] == '>' || namedata[i] == '|' || namedata[i] == '/' || namedata[i] == '\\' || namedata[i] == ':' || namedata[i] == '?' || namedata[i] == '"' || namedata[i] == '*' || namedata[i] == '%' {
			namedata = namedata[:i] + "%" + fmt.Sprintf("%02x", namedata[i]) + namedata[i+1:]
		}
	}
	if len(namedata) > 0 && namedata[0] == '.' {
		namedata = "%" + fmt.Sprintf("%02x", namedata[0]) + namedata[1:]
	}
	if len(namedata) > 0 && namedata[len(namedata)-1] == '.' {
		namedata = namedata[:len(namedata)-1] + "%" + fmt.Sprintf("%02x", namedata[0])
	}
	return namedata
}

func UrlDataDecode(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '%' {
			val, vale := strconv.ParseInt(name[i : i+3][1:], 16, 32)
			if vale == nil {
				name = name[:i] + string([]byte{byte(val)}) + name[i+3:]
			}
		}
	}
	return name
}

func UrlDataEncode(namedata string) string {
	for i := len(namedata) - 1; i >= 0; i-- {
		if namedata[i] >= 0 && namedata[i] <= 32 || namedata[i] >= 128 || namedata[i] == '/' || namedata[i] == '\\' || namedata[i] == ':' || namedata[i] == '?' || namedata[i] == '"' || namedata[i] == '%' || namedata[i] == '&' || namedata[i] == '=' || namedata[i] == '$' || namedata[i] == '#' || namedata[i] == '^' || namedata[i] == '{' || namedata[i] == '}' {
			namedata = namedata[:i] + "%" + fmt.Sprintf("%02x", namedata[i]) + namedata[i+1:]
		}
	}
	return namedata
}

func StringToMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str)) // 需要加密的字符串为 123456
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr) // 输出加密结果
}

func ReadFile(filepath string) ([]byte, error) {
	ff, ffe := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if ffe != nil {
		return nil, ffe
	}
	defer ff.Close()
	endpos, endpose := ff.Seek(0, os.SEEK_END)
	if endpose != nil {
		return nil, ffe
	}
	startpos, startpose := ff.Seek(0, os.SEEK_SET)
	if startpose != nil || startpos != 0 {
		return nil, ffe
	}
	buf := make([]byte, endpos)
	rcnt, err := io.ReadFull(ff, buf)
	if err != nil || int64(rcnt) != endpos {
		return nil, err
	}
	return buf, nil
}

func Count(cursegment, separator []byte) (count int) {
	for i := 0; i < len(cursegment); {
		if bytes.Compare(cursegment[i:i+len(separator)], separator) == 0 {
			count += 1
			i += len(separator)
		} else {
			i += 1
		}
	}
	return count
}

func Truncate(cursegment, separator []byte, n int) (trunc []byte) {
	for i := 0; i < len(cursegment); {
		if bytes.Compare(cursegment[i:i+len(separator)], separator) == 0 {
			n -= 1
			if n == 0 {
				trunc = cursegment[:i+len(separator)]
				return trunc
			}
			i += len(separator)
		} else {
			i += 1
		}
	}
	return []byte{}
}

func SaveFile(filepath string, data []byte) error {
	return WriteFile(filepath, data)
}

func WriteFile(filepath string, data []byte) error {
	//ff, ffe := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0666)
	ff, ffe := os.Create(filepath)
	if ffe != nil {
		return ffe
	}
	defer ff.Close()
	_, we := ff.Write(data)
	if we != nil {
		return we
	}
	synce := ff.Sync()
	if synce != nil {
		return synce
	}
	ff.Close()
	return nil
}

func AppendFile(filepath string, data []byte, accesswright os.FileMode) error {
	//ff, ffe := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0666)
	ff, ffe := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, accesswright)
	if ffe != nil {
		return ffe
	}
	defer ff.Close()
	_, e2 := ff.Seek(0, os.SEEK_END)
	if e2 != nil {
		return e2
	}
	_, e := ff.Write(data)
	if e != nil {
		return e
	}
	return nil
}

func StringBytesMapToBytes(mdata sync.Map) (outbt []byte) {
	keybt := make([]byte, 4)
	vallenbt := make([]byte, 4)
	var keyint, vallen uint32
	mdata.Range(func(key, value interface{}) bool {
		keyint = uint32(len(key.(string)))
		binary.BigEndian.PutUint32(keybt, keyint)
		vallen = uint32(len(value.([]byte)))
		binary.BigEndian.PutUint32(vallenbt, vallen)
		outbt = append(outbt, keybt...)
		outbt = append(outbt, []byte(key.(string))...)
		outbt = append(outbt, vallenbt...)
		outbt = append(outbt, value.([]byte)...)
		return true
	})
	return outbt
}

func StringBytesMapFromBytes(mdatastr []byte) (mdata sync.Map) {
	var keylen, vallen, startpos uint32
	for uint32(startpos)+8 <= uint32(len(mdatastr)) {
		keylen = binary.BigEndian.Uint32(mdatastr[startpos : startpos+4])
		keybt := mdatastr[startpos+4 : startpos+4+keylen]
		vallen = binary.BigEndian.Uint32(mdatastr[startpos+4+keylen : startpos+4+keylen+4])
		mdata.Store(string(keybt), mdatastr[startpos+4+keylen+4:startpos+4+keylen+4+vallen])
		startpos += 4 + keylen + 4 + vallen
		if startpos >= uint32(len(mdatastr)) {
			break
		}
	}
	return mdata
}

func RandAlpha(min, max int) (outbt []byte) {
	sublen := min + rand.Intn(max-min+1)
	for i := 0; i < sublen; i++ {
		outbt = append(outbt, byte(0x41+rand.Intn(26)))
	}
	return outbt
}

func RandPrintChar(min, max int) (outbt []byte) {
	sublen := min + rand.Intn(max-min+1)
	for i := 0; i < sublen; i++ {
		outbt = append(outbt, byte(0x20+rand.Intn(0x7e-0x20)))
	}
	return outbt
}

func RandBase64Charset(min, max int) (outbt []byte) {
	sublen := min + rand.Intn(max-min+1)
	for i := 0; i < sublen; i++ {
		numrnd := rand.Intn(52)
		if numrnd < 26 {
			outbt = append(outbt, byte(0x41+rand.Intn(26)))
		} else if numrnd >= 26 && numrnd < 36 {
			outbt = append(outbt, byte(0x30+rand.Intn(10)))
		} else {
			outbt = append(outbt, byte(0x61+rand.Intn(26)))
		}
	}
	return outbt
}

func RandUpperNum(min, max int) (outbt []byte) {
	sublen := min + rand.Intn(max-min+1)
	for i := 0; i < sublen; i++ {
		if rand.Intn(36) < 26 {
			outbt = append(outbt, byte(0x41+rand.Intn(26)))
		} else {
			outbt = append(outbt, byte(0x30+rand.Intn(10)))
		}
	}
	return outbt
}

func RandLowerNum(min, max int) (outbt []byte) {
	sublen := min + rand.Intn(max-min+1)
	for i := 0; i < sublen; i++ {
		if rand.Intn(36) < 26 {
			outbt = append(outbt, byte(0x61+rand.Intn(26)))
		} else {
			outbt = append(outbt, byte(0x30+rand.Intn(10)))
		}
	}
	return outbt
}

func RandHex(min, max int) (outbt []byte) {
	sublen := min + rand.Intn(max-min+1)
	for i := 0; i < sublen; i++ {
		if rand.Intn(16) < 6 {
			outbt = append(outbt, byte(0x61+rand.Intn(6)))
		} else {
			outbt = append(outbt, byte(0x30+rand.Intn(10)))
		}
	}
	return outbt
}

func UniqueAdd(set interface{}, val interface{}) interface{} {
	switch set.(type) {
	case [][]byte:
		{
			for i := len(set.([][]byte)) - 1; i >= 0; i-- {
				if Byte1DCompare(set.([][]byte)[i], val.([]byte)) {
					return set
				}
			}
			return append(set.([][]byte), val.([]byte))
		}
	case [][]string:
		{
			for i := len(set.([][]string)) - 1; i >= 0; i-- {
				if String1DCompare(set.([][]string)[i], val.([]string)) {
					return set
				}
			}
			return append(set.([][]string), val.([]string))
		}
	case []int:
		{
			for i := len(set.([]int)) - 1; i >= 0; i-- {
				if set.([]int)[i] == val.(int) {
					return set
				}
			}
			return append(set.([]int), val.(int))
		}
	case []int8:
		{
			for i := len(set.([]int8)) - 1; i >= 0; i-- {
				if set.([]int8)[i] == val.(int8) {
					return set
				}
			}
			return append(set.([]int8), val.(int8))
		}
	case []int16:
		{
			for i := len(set.([]int16)) - 1; i >= 0; i-- {
				if set.([]int16)[i] == val.(int16) {
					return set
				}
			}
			return append(set.([]int16), val.(int16))
		}
	case []int32:
		{
			for i := len(set.([]int32)) - 1; i >= 0; i-- {
				if set.([]int32)[i] == val.(int32) {
					return set
				}
			}
			return append(set.([]int32), val.(int32))
		}
	case []int64:
		{
			for i := len(set.([]int64)) - 1; i >= 0; i-- {
				if set.([]int64)[i] == val.(int64) {
					return set
				}
			}
			return append(set.([]int64), val.(int64))
		}
	case []uint8:
		{
			for i := len(set.([]uint8)) - 1; i >= 0; i-- {
				if set.([]uint8)[i] == val.(uint8) {
					return set
				}
			}
			return append(set.([]uint8), val.(uint8))
		}
	case []uint16:
		{
			for i := len(set.([]uint16)) - 1; i >= 0; i-- {
				if set.([]uint16)[i] == val.(uint16) {
					return set
				}
			}
			return append(set.([]uint16), val.(uint16))

		}
	case []uint32:
		{
			for i := len(set.([]uint32)) - 1; i >= 0; i-- {
				if set.([]uint32)[i] == val.(uint32) {
					return set
				}
			}
			return append(set.([]uint32), val.(uint32))

		}
	case []uint64:
		{
			for i := len(set.([]uint64)) - 1; i >= 0; i-- {
				if set.([]uint64)[i] == val.(uint64) {
					return set
				}
			}
			return append(set.([]uint64), val.(uint64))
		}
	case []string:
		{
			for i := len(set.([]string)) - 1; i >= 0; i-- {
				if set.([]string)[i] == val.(string) {
					return set
				}
			}
			return append(set.([]string), val.(string))
		}
	default:
		panic("error")
	}
	return set
}

func RemoveOne(set interface{}, val interface{}) interface{} {
	switch set.(type) {
	case [][]byte:
		{
			for i := len(set.([][]byte)) - 1; i >= 0; i-- {
				if Byte1DCompare(set.([][]byte)[i], val.([]byte)) {
					set = append(set.([][]byte)[:i], set.([][]byte)[i+1:]...)
					return set
				}
			}
		}
	case [][]string:
		{
			for i := len(set.([][]string)) - 1; i >= 0; i-- {
				if String1DCompare(set.([][]string)[i], val.([]string)) {
					set = append(set.([][]string)[:i], set.([][]string)[i+1:]...)
					return set
				}
			}
		}
	case []int:
		{
			for i := len(set.([]int)) - 1; i >= 0; i-- {
				if set.([]int)[i] == val.(int) {
					set = append(set.([]int)[:i], set.([]int)[i+1:]...)
					return set
				}
			}
		}
	case []int8:
		{
			for i := len(set.([]int8)) - 1; i >= 0; i-- {
				if set.([]int8)[i] == val.(int8) {
					set = append(set.([]int8)[:i], set.([]int8)[i+1:]...)
					return set
				}
			}
		}
	case []int16:
		{
			for i := len(set.([]int16)) - 1; i >= 0; i-- {
				if set.([]int16)[i] == val.(int16) {
					set = append(set.([]int16)[:i], set.([]int16)[i+1:]...)
					return set
				}
			}
		}
	case []int32:
		{
			for i := len(set.([]int32)) - 1; i >= 0; i-- {
				if set.([]int32)[i] == val.(int32) {
					set = append(set.([]int32)[:i], set.([]int32)[i+1:]...)
					return set
				}
			}
		}
	case []int64:
		{
			for i := len(set.([]int64)) - 1; i >= 0; i-- {
				if set.([]int64)[i] == val.(int64) {
					set = append(set.([]int64)[:i], set.([]int64)[i+1:]...)
					return set
				}
			}
		}
	case []uint8:
		{
			for i := len(set.([]uint8)) - 1; i >= 0; i-- {
				if set.([]uint8)[i] == val.(uint8) {
					set = append(set.([]uint8)[:i], set.([]uint8)[i+1:]...)
					return set
				}
			}
		}
	case []uint16:
		{
			for i := len(set.([]uint16)) - 1; i >= 0; i-- {
				if set.([]uint16)[i] == val.(uint16) {
					set = append(set.([]uint16)[:i], set.([]uint16)[i+1:]...)
					return set
				}
			}

		}
	case []uint32:
		{
			for i := len(set.([]uint32)) - 1; i >= 0; i-- {
				if set.([]uint32)[i] == val.(uint32) {
					set = append(set.([]uint32)[:i], set.([]uint32)[i+1:]...)
					return set
				}
			}

		}
	case []uint64:
		{
			for i := len(set.([]uint64)) - 1; i >= 0; i-- {
				if set.([]uint64)[i] == val.(uint64) {
					set = append(set.([]uint64)[:i], set.([]uint64)[i+1:]...)
					return set
				}
			}
		}
	case []string:
		{
			for i := len(set.([]string)) - 1; i >= 0; i-- {
				if set.([]string)[i] == val.(string) {
					set = append(set.([]string)[:i], set.([]string)[i+1:]...)
					return set
				}
			}
		}
	default:
		panic("error")
	}
	return set
}

func IdsToBytes32(ids []uint32) (outbt []byte) {
	idbt := make([]byte, 4)
	for _, id := range ids {
		binary.BigEndian.PutUint32(idbt, id)
		outbt = append(outbt, idbt...)
	}
	return outbt
}

func BytesToIds32(outbt []byte) (ids []uint32) {
	i := 0
	for i < len(outbt) {
		ids = append(ids, binary.BigEndian.Uint32(outbt[i:i+4]))
		i += 4
	}
	return ids
}

func IdsToBytes64(ids []uint64) (outbt []byte) {
	idbt := make([]byte, 8)
	for _, id := range ids {
		binary.BigEndian.PutUint64(idbt, id)
		outbt = append(outbt, idbt...)
	}
	return outbt
}

func BytesToIds64(outbt []byte) (ids []uint64) {
	i := 0
	for i < len(outbt) {
		ids = append(ids, binary.BigEndian.Uint64(outbt[i:i+8]))
		i += 8
	}
	return ids
}

//src max size 32765
func FlateEncode(buf, src []byte, level int) []byte {
	var buf2 []byte
	if buf == nil {
	} else {
		buf2 = buf
	}
	b := bytes.NewBuffer(buf2)
	b.Reset()
	zw, err := flate.NewWriter(b, level)
	if err != nil {
		panic("FlateEncode Error 2460.")
	}
	zw.Write(src)
	zw.Close()
	if buf != nil {
		return buf[:b.Len()]
	} else {
		outbt := b.Bytes()
		if cap(outbt) > len(outbt)+4 {
			binary.BigEndian.PutUint32(outbt[b.Len():b.Len()+4], uint32(len(src)))
			return outbt[:b.Len()+4]
		} else {
			lbt := make([]byte, 4)
			binary.BigEndian.PutUint32(lbt, uint32(len(src)))
			return append(outbt, lbt...)
		}
	}
}

//src max size 32765
func FlateDecode(outbuf, src []byte) []byte {
	if src == nil || len(src) == 0 {
		return []byte{}
	}
	if outbuf == nil {
		b := bytes.NewBuffer(src[:len(src)-4])
		zr := flate.NewReader(nil)
		if err := zr.(flate.Resetter).Reset(b, nil); err != nil {
			fmt.Println(GetGoRoutineID(), " FlateDecode src error. src:", src)
			panic("FlateDecode error 2471.")
		}
		oldlen := binary.BigEndian.Uint32(src[len(src)-4:])
		b2buf := make([]byte, oldlen)
		zr.Read(b2buf)
		if err := zr.Close(); err != nil {
			fmt.Println(GetGoRoutineID(), " FlateDecode src error. src:", src)
			panic("FlateDecode error 2481.")
		}
		return b2buf
	} else {
		b := bytes.NewBuffer(src)
		zr := flate.NewReader(nil)
		if err := zr.(flate.Resetter).Reset(b, nil); err != nil {
			fmt.Println(GetGoRoutineID(), " FlateDecode src error. src:", src)
			panic("FlateDecode error 2471.")
		}
		b2 := bytes.NewBuffer(outbuf)
		b2.Reset()
		outw := bufio.NewWriter(b2)
		if _, err := io.Copy(outw, zr); err != nil {
			fmt.Println(GetGoRoutineID(), " FlateDecode src error. src:", src)
			panic("DelateDecode Error 2478.")
		}
		if err := zr.Close(); err != nil {
			fmt.Println(GetGoRoutineID(), " FlateDecode src error. src:", src)
			panic("FlateDecode error 2481.")
		}
		return outbuf[:b2.Len()]
	}
}

func InterfaceIntItemQuickSortAsc(arr []interface{}, start, end, setlength, insetpos int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[((start+end)/2)*setlength+insetpos].(int)
		for i <= j {
			for arr[(i)*setlength+insetpos].(int) < key {
				i++
			}
			for arr[(j)*setlength+insetpos].(int) > key {
				j--
			}
			if i <= j {
				for k := 0; k < setlength; k++ {
					arr[i*setlength+k], arr[j*setlength+k] = arr[j*setlength+k], arr[i*setlength+k]
				}
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			InterfaceIntItemQuickSortAsc(arr, start, j, setlength, insetpos)
		}
		if end > i {
			InterfaceIntItemQuickSortAsc(arr, i, end, setlength, insetpos)
		}
	}
	return bexchange
}

func IntQuickSortAsc(arr []int, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			IntQuickSortAsc(arr, start, j)
		}
		if end > i {
			IntQuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Uint8QuickSortAsc(arr []uint8, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint8QuickSortAsc(arr, start, j)
		}
		if end > i {
			Uint8QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Uint16QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint16QuickSortAsc(arr, start, j)
		}
		if end > i {
			Uint16QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Uint32QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint32QuickSortAsc(arr, start, j)
		}
		if end > i {
			Uint32QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Uint64QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint64QuickSortAsc(arr, start, j)
		}
		if end > i {
			Uint64QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Int64QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int64QuickSortAsc(arr, start, j)
		}
		if end > i {
			Int64QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Int32QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int32QuickSortAsc(arr, start, j)
		}
		if end > i {
			Int32QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Int16QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int16QuickSortAsc(arr, start, j)
		}
		if end > i {
			Int16QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func Int8QuickSortAsc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int8QuickSortAsc(arr, start, j)
		}
		if end > i {
			Int8QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func StringQuickSortAsc(arr []string, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for strings.Compare(arr[i], key) < 0 {
				i++
			}
			for strings.Compare(arr[j], key) > 0 {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			StringQuickSortAsc(arr, start, j)
		}
		if end > i {
			StringQuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func StringIsFloat64QuickSortAsc(arr []string, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := Float32FromStr(arr[(start+end)/2])
		for i <= j {
			for Float32FromStr(arr[i]) < key {
				i++
			}
			for Float32FromStr(arr[j]) > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			StringIsFloat64QuickSortAsc(arr, start, j)
		}
		if end > i {
			StringIsFloat64QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func BytesQuickSortAsc(arr [][]byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for bytes.Compare(arr[i], key) < 0 {
				i++
			}
			for bytes.Compare(arr[j], key) > 0 {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			BytesQuickSortAsc(arr, start, j)
		}
		if end > i {
			BytesQuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func InterfaceIntQuickSortDesc(arr []interface{}, start, end, setlength, insetpos int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[((start+end)/2)*setlength+insetpos].(int)
		for i <= j {
			for arr[i*setlength+insetpos].(int) > key {
				i++
			}
			for arr[j*setlength+insetpos].(int) < key {
				j--
			}
			if i <= j {
				for k := 0; k < setlength; k++ {
					arr[i*setlength+k], arr[j*setlength+k] = arr[j*setlength+k], arr[i*setlength+k]
				}
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			InterfaceIntQuickSortDesc(arr, start, j, setlength, insetpos)
		}
		if end > i {
			InterfaceIntQuickSortDesc(arr, i, end, setlength, insetpos)
		}
	}
	return bexchange
}

func IntQuickSortDesc(arr []int, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			IntQuickSortDesc(arr, start, j)
		}
		if end > i {
			IntQuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Uint8QuickSortDesc(arr []uint8, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint8QuickSortDesc(arr, start, j)
		}
		if end > i {
			Uint8QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Uint16QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint16QuickSortDesc(arr, start, j)
		}
		if end > i {
			Uint16QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Uint32QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint32QuickSortDesc(arr, start, j)
		}
		if end > i {
			Uint32QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Uint64QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Uint64QuickSortDesc(arr, start, j)
		}
		if end > i {
			Uint64QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Int64QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int64QuickSortDesc(arr, start, j)
		}
		if end > i {
			Int64QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Int32QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int32QuickSortDesc(arr, start, j)
		}
		if end > i {
			Int32QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Int16QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int16QuickSortDesc(arr, start, j)
		}
		if end > i {
			Int16QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func Int8QuickSortDesc(arr []uint16, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] > key {
				i++
			}
			for arr[j] < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			Int8QuickSortDesc(arr, start, j)
		}
		if end > i {
			Int8QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func StringQuickSortDesc(arr []string, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for strings.Compare(arr[i], key) > 0 {
				i++
			}
			for strings.Compare(arr[j], key) < 0 {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			StringQuickSortDesc(arr, start, j)
		}
		if end > i {
			StringQuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func StringIsFloat64QuickSortDesc(arr []string, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := Float64FromStr(arr[(start+end)/2])
		for i <= j {
			for Float64FromStr(arr[i]) > key {
				i++
			}
			for Float64FromStr(arr[j]) < key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			StringIsFloat64QuickSortDesc(arr, start, j)
		}
		if end > i {
			StringIsFloat64QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func BytesQuickSortDesc(arr [][]byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for bytes.Compare(arr[i], key) > 0 {
				i++
			}
			for bytes.Compare(arr[j], key) < 0 {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			BytesQuickSortDesc(arr, start, j)
		}
		if end > i {
			BytesQuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteUint16QuickSortAsc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := binary.BigEndian.Uint16(arr[(start+end)/2*2 : (start+end)/2*2+2])
		for i <= j {
			for binary.BigEndian.Uint16(arr[i*2:i*2+2]) < key {
				i++
			}
			for binary.BigEndian.Uint16(arr[j*2:j*2+2]) > key {
				j--
			}
			if i <= j {
				arr[i*2], arr[j*2] = arr[j*2], arr[i*2]
				arr[i*2+1], arr[j*2+1] = arr[j*2+1], arr[i*2+1]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint16QuickSortAsc(arr, start, j)
		}
		if end > i {
			ByteUint16QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

//return value meat to have exchange
func ByteUint32QuickSortAsc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := binary.BigEndian.Uint32(arr[(start+end)/2*4 : (start+end)/2*4+4])
		for i <= j {
			for binary.BigEndian.Uint32(arr[i*4:i*4+4]) < key {
				i++
			}
			for binary.BigEndian.Uint32(arr[j*4:j*4+4]) > key {
				j--
			}
			if i <= j {
				arr[i*4], arr[j*4] = arr[j*4], arr[i*4]
				arr[i*4+1], arr[j*4+1] = arr[j*4+1], arr[i*4+1]
				arr[i*4+2], arr[j*4+2] = arr[j*4+2], arr[i*4+2]
				arr[i*4+3], arr[j*4+3] = arr[j*4+3], arr[i*4+3]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint32QuickSortAsc(arr, start, j)
		}
		if end > i {
			ByteUint32QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func ByteUint64QuickSortAsc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := binary.BigEndian.Uint64(arr[(start+end)/2*8 : (start+end)/2*8+8])
		for i <= j {
			for binary.BigEndian.Uint64(arr[i*8:i*8+8]) < key {
				i++
			}
			for binary.BigEndian.Uint64(arr[j*8:j*8+8]) > key {
				j--
			}
			if i <= j {
				arr[i*8], arr[j*8] = arr[j*8], arr[i*8]
				arr[i*8+1], arr[j*8+1] = arr[j*8+1], arr[i*8+1]
				arr[i*8+2], arr[j*8+2] = arr[j*8+2], arr[i*8+2]
				arr[i*8+3], arr[j*8+3] = arr[j*8+3], arr[i*8+3]
				arr[i*8+4], arr[j*8+4] = arr[j*8+4], arr[i*8+4]
				arr[i*8+5], arr[j*8+5] = arr[j*8+5], arr[i*8+5]
				arr[i*8+6], arr[j*8+6] = arr[j*8+6], arr[i*8+6]
				arr[i*8+7], arr[j*8+7] = arr[j*8+7], arr[i*8+7]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint64QuickSortAsc(arr, start, j)
		}
		if end > i {
			ByteUint64QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func ByteInt16QuickSortAsc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := int16(binary.BigEndian.Uint16(arr[(start+end)/2*2 : (start+end)/2*2+2]))
		for i <= j {
			for int16(binary.BigEndian.Uint16(arr[i*2:i*2+2])) < key {
				i++
			}
			for int16(binary.BigEndian.Uint16(arr[j*2:j*2+2])) > key {
				j--
			}
			if i <= j {
				arr[i*2], arr[j*2] = arr[j*2], arr[i*2]
				arr[i*2+1], arr[j*2+1] = arr[j*2+1], arr[i*2+1]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteInt16QuickSortAsc(arr, start, j)
		}
		if end > i {
			ByteInt16QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func ByteInt32QuickSortAsc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := int32(binary.BigEndian.Uint32(arr[(start+end)/2*4 : (start+end)/2*4+4]))
		for i <= j {
			for int32(binary.BigEndian.Uint32(arr[i*4:i*4+4])) < key {
				i++
			}
			for int32(binary.BigEndian.Uint32(arr[j*4:j*4+4])) > key {
				j--
			}
			if i <= j {
				arr[i*4], arr[j*4] = arr[j*4], arr[i*4]
				arr[i*4+1], arr[j*4+1] = arr[j*4+1], arr[i*4+1]
				arr[i*4+2], arr[j*4+2] = arr[j*4+2], arr[i*4+2]
				arr[i*4+3], arr[j*4+3] = arr[j*4+3], arr[i*4+3]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteInt32QuickSortAsc(arr, start, j)
		}
		if end > i {
			ByteInt32QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func ByteInt64QuickSortAsc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := int64(binary.BigEndian.Uint64(arr[(start+end)/2*8 : (start+end)/2*8+8]))
		for i <= j {
			for int64(binary.BigEndian.Uint64(arr[i*8:i*8+8])) < key {
				i++
			}
			for int64(binary.BigEndian.Uint64(arr[j*8:j*8+8])) > key {
				j--
			}
			if i <= j {
				arr[i*8], arr[j*8] = arr[j*8], arr[i*8]
				arr[i*8+1], arr[j*8+1] = arr[j*8+1], arr[i*8+1]
				arr[i*8+2], arr[j*8+2] = arr[j*8+2], arr[i*8+2]
				arr[i*8+3], arr[j*8+3] = arr[j*8+3], arr[i*8+3]
				arr[i*8+4], arr[j*8+4] = arr[j*8+4], arr[i*8+4]
				arr[i*8+5], arr[j*8+5] = arr[j*8+5], arr[i*8+5]
				arr[i*8+6], arr[j*8+6] = arr[j*8+6], arr[i*8+6]
				arr[i*8+7], arr[j*8+7] = arr[j*8+7], arr[i*8+7]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint64QuickSortAsc(arr, start, j)
		}
		if end > i {
			ByteUint64QuickSortAsc(arr, i, end)
		}
	}
	return bexchange
}

func ByteUint16QuickSortDesc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := binary.BigEndian.Uint16(arr[(start+end)/2*2 : (start+end)/2*2+2])
		for i <= j {
			for binary.BigEndian.Uint16(arr[i*2:i*2+2]) > key {
				i++
			}
			for binary.BigEndian.Uint16(arr[j*2:j*2+2]) < key {
				j--
			}
			if i <= j {
				arr[i*2], arr[j*2] = arr[j*2], arr[i*2]
				arr[i*2+1], arr[j*2+1] = arr[j*2+1], arr[i*2+1]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint16QuickSortDesc(arr, start, j)
		}
		if end > i {
			ByteUint16QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteUint32QuickSortDesc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := binary.BigEndian.Uint32(arr[(start+end)/2*4 : (start+end)/2*4+4])
		for i <= j {
			for binary.BigEndian.Uint32(arr[i*4:i*4+4]) > key {
				i++
			}
			for binary.BigEndian.Uint32(arr[j*4:j*4+4]) < key {
				j--
			}
			if i <= j {
				arr[i*4], arr[j*4] = arr[j*4], arr[i*4]
				arr[i*4+1], arr[j*4+1] = arr[j*4+1], arr[i*4+1]
				arr[i*4+2], arr[j*4+2] = arr[j*4+2], arr[i*4+2]
				arr[i*4+3], arr[j*4+3] = arr[j*4+3], arr[i*4+3]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint32QuickSortDesc(arr, start, j)
		}
		if end > i {
			ByteUint32QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteUint64QuickSortDesc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := binary.BigEndian.Uint64(arr[(start+end)/2*8 : (start+end)/2*8+8])
		for i <= j {
			for binary.BigEndian.Uint64(arr[i*8:i*8+8]) > key {
				i++
			}
			for binary.BigEndian.Uint64(arr[j*8:j*8+8]) < key {
				j--
			}
			if i <= j {
				arr[i*8], arr[j*8] = arr[j*8], arr[i*8]
				arr[i*8+1], arr[j*8+1] = arr[j*8+1], arr[i*8+1]
				arr[i*8+2], arr[j*8+2] = arr[j*8+2], arr[i*8+2]
				arr[i*8+3], arr[j*8+3] = arr[j*8+3], arr[i*8+3]
				arr[i*8+4], arr[j*8+4] = arr[j*8+4], arr[i*8+4]
				arr[i*8+5], arr[j*8+5] = arr[j*8+5], arr[i*8+5]
				arr[i*8+6], arr[j*8+6] = arr[j*8+6], arr[i*8+6]
				arr[i*8+7], arr[j*8+7] = arr[j*8+7], arr[i*8+7]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint64QuickSortDesc(arr, start, j)
		}
		if end > i {
			ByteUint64QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteInt16QuickSortDesc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := int16(binary.BigEndian.Uint16(arr[(start+end)/2*2 : (start+end)/2*2+2]))
		for i <= j {
			for int16(binary.BigEndian.Uint16(arr[i*2:i*2+2])) > key {
				i++
			}
			for int16(binary.BigEndian.Uint16(arr[j*2:j*2+2])) < key {
				j--
			}
			if i <= j {
				arr[i*2], arr[j*2] = arr[j*2], arr[i*2]
				arr[i*2+1], arr[j*2+1] = arr[j*2+1], arr[i*2+1]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteInt16QuickSortDesc(arr, start, j)
		}
		if end > i {
			ByteInt16QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteInt32QuickSortDesc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := int32(binary.BigEndian.Uint32(arr[(start+end)/2*4 : (start+end)/2*4+4]))
		for i <= j {
			for int32(binary.BigEndian.Uint32(arr[i*4:i*4+4])) > key {
				i++
			}
			for int32(binary.BigEndian.Uint32(arr[j*4:j*4+4])) < key {
				j--
			}
			if i <= j {
				arr[i*4], arr[j*4] = arr[j*4], arr[i*4]
				arr[i*4+1], arr[j*4+1] = arr[j*4+1], arr[i*4+1]
				arr[i*4+2], arr[j*4+2] = arr[j*4+2], arr[i*4+2]
				arr[i*4+3], arr[j*4+3] = arr[j*4+3], arr[i*4+3]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteInt32QuickSortDesc(arr, start, j)
		}
		if end > i {
			ByteInt32QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteInt64QuickSortDesc(arr []byte, start, end int) (bexchange bool) {
	if start < end {
		i, j := start, end
		key := int64(binary.BigEndian.Uint64(arr[(start+end)/2*8 : (start+end)/2*8+8]))
		for i <= j {
			for int64(binary.BigEndian.Uint64(arr[i*8:i*8+8])) > key {
				i++
			}
			for int64(binary.BigEndian.Uint64(arr[j*8:j*8+8])) < key {
				j--
			}
			if i <= j {
				arr[i*8], arr[j*8] = arr[j*8], arr[i*8]
				arr[i*8+1], arr[j*8+1] = arr[j*8+1], arr[i*8+1]
				arr[i*8+2], arr[j*8+2] = arr[j*8+2], arr[i*8+2]
				arr[i*8+3], arr[j*8+3] = arr[j*8+3], arr[i*8+3]
				arr[i*8+4], arr[j*8+4] = arr[j*8+4], arr[i*8+4]
				arr[i*8+5], arr[j*8+5] = arr[j*8+5], arr[i*8+5]
				arr[i*8+6], arr[j*8+6] = arr[j*8+6], arr[i*8+6]
				arr[i*8+7], arr[j*8+7] = arr[j*8+7], arr[i*8+7]
				i++
				j--
				bexchange = true
			}
		}

		if start < j {
			ByteUint64QuickSortDesc(arr, start, j)
		}
		if end > i {
			ByteUint64QuickSortDesc(arr, i, end)
		}
	}
	return bexchange
}

func ByteUint16RemoveRepeat(arr []byte) (arrout []byte) {
	cnt := len(arr) / 2
	for i := cnt - 2; i >= 0; i-- {
		if bytes.Compare(arr[i*2:i*2+2], arr[(i+1)*2:(i+1)*2+2]) == 0 {
			arr = append(arr[:i*2], arr[(i+1)*2:]...)
		}
	}
	return arr
}

func ByteUint32RemoveRepeat(arr []byte) (arrout []byte) {
	cnt := len(arr) / 4
	for i := cnt - 2; i >= 0; i-- {
		if bytes.Compare(arr[i*4:i*4+4], arr[(i+1)*4:(i+1)*4+4]) == 0 {
			arr = append(arr[:i*4], arr[(i+1)*4:]...)
		}
	}
	return arr
}

func ByteUint64RemoveRepeat(arr []byte) (arrout []byte) {
	cnt := len(arr) / 8
	for i := cnt - 2; i >= 0; i-- {
		if bytes.Compare(arr[i*8:i*8+8], arr[(i+1)*8:(i+1)*8+8]) == 0 {
			arr = append(arr[:i*8], arr[(i+1)*8:]...)
		}
	}
	return arr
}

func QuickIndex(sortvalls, value interface{}) (index int) {
	index = -1
	switch sortvalls.(type) {
	case []int:
		start := 0
		end := len(sortvalls.([]int)) - 1
		cur := end / 2
		for true {
			if value.(int) < sortvalls.([]int)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int) > sortvalls.([]int)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []int8:
		start := 0
		end := len(sortvalls.([]int8)) - 1
		cur := end / 2
		for true {
			if value.(int8) < sortvalls.([]int8)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int8) > sortvalls.([]int8)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []int16:
		start := 0
		end := len(sortvalls.([]int16)) - 1
		cur := end / 2
		for true {
			if value.(int16) < sortvalls.([]int16)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int16) > sortvalls.([]int16)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []int32:
		start := 0
		end := len(sortvalls.([]int32)) - 1
		cur := end / 2
		for true {
			if value.(int32) < sortvalls.([]int32)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int32) > sortvalls.([]int32)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []int64:
		start := 0
		end := len(sortvalls.([]int64)) - 1
		cur := end / 2
		for true {
			if value.(int64) < sortvalls.([]int64)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int64) > sortvalls.([]int64)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []uint8:
		start := 0
		end := len(sortvalls.([]uint8)) - 1
		cur := end / 2
		for true {
			if value.(uint8) < sortvalls.([]uint8)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint8) > sortvalls.([]uint8)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []uint16:
		start := 0
		end := len(sortvalls.([]uint16)) - 1
		cur := end / 2
		for true {
			if value.(uint16) < sortvalls.([]uint16)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint16) > sortvalls.([]uint16)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []uint32:
		start := 0
		end := len(sortvalls.([]uint32)) - 1
		cur := end / 2
		for true {
			if value.(uint32) < sortvalls.([]uint32)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint32) > sortvalls.([]uint32)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []uint64:
		start := 0
		end := len(sortvalls.([]uint64)) - 1
		cur := end / 2
		for true {
			if value.(uint64) < sortvalls.([]uint64)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint64) > sortvalls.([]uint64)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []float32:
		start := 0
		end := len(sortvalls.([]float32)) - 1
		cur := end / 2
		for true {
			if value.(float32) < sortvalls.([]float32)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(float32) > sortvalls.([]float32)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []float64:
		start := 0
		end := len(sortvalls.([]float64)) - 1
		cur := end / 2
		for true {
			if value.(float64) < sortvalls.([]float64)[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(float64) > sortvalls.([]float64)[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	default:
		panic("unsupport error")
	}
	return -1
}

func ByteQuickIndex(sortvalls []byte, value interface{}) (index int) {
	index = -1
	switch value.(type) {
	case []int:
		panic("not support error.")
	case []int8:
		start := 0
		end := len(sortvalls) - 1
		cur := end / 2
		for true {
			if value.(int8) < int8(sortvalls[cur]) {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int8) > int8(sortvalls[cur]) {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []int16:
		start := 0
		end := len(sortvalls)/2 - 1
		cur := end / 2
		for true {
			curval := int16(binary.BigEndian.Uint16(sortvalls[2*cur : 2*cur+2]))
			if value.(int16) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int16) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []int32:
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := int32(binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4]))
			if value.(int32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []int64:
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := int64(binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8]))
			if value.(int64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(int64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []uint8:
		start := 0
		end := len(sortvalls) - 1
		cur := end / 2
		for true {
			if value.(uint8) < sortvalls[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint8) > sortvalls[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return cur
			}
		}
	case []uint16:
		start := 0
		end := len(sortvalls)/2 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint16(sortvalls[2*cur : 2*cur+2])
			if value.(uint16) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint16) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []uint32:
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4])
			if value.(uint32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []uint64:
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8])
			if value.(uint64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(uint64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []float32:
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := math.Float32frombits(binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4]))
			if value.(float32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(float32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	case []float64:
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := math.Float64frombits(binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8]))
			if value.(float64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else if value.(float64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return -1
					}
				} else {
					cur = cur2
				}
			} else {
				return cur
			}
		}
	default:
		panic("unsupport error")
	}
	return -1
}

func ByteQuickInsert(sortvalls []byte, value interface{}) (outbt []byte) {
	switch value.(type) {
	case []int:
		panic("not support error.")
	case []int8:
		if len(sortvalls) == 0 {
			return []byte{byte(value.(int8))}
		}
		start := 0
		end := len(sortvalls) - 1
		cur := end / 2
		for true {
			if value.(int8) < int8(sortvalls[cur]) {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*1+1], sortvalls[cur*1:]...)
						sortvalls[cur*4] = byte(value.(int8))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int8) > int8(sortvalls[cur]) {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*1+1], sortvalls[(cur+1)*1:]...)
						sortvalls[(cur+1)*1] = byte(value.(int8))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return sortvalls
			}
		}
	case []int16:
		if len(sortvalls) == 0 {
			return PackUint16(uint16(value.(int16)))
		}
		start := 0
		end := len(sortvalls)/2 - 1
		cur := end / 2
		for true {
			curval := int16(binary.BigEndian.Uint16(sortvalls[2*cur : 2*cur+2]))
			if value.(int16) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*2+2], sortvalls[cur*2:]...)
						copy(sortvalls[cur*2:cur*2+2], PackUint16(uint16(value.(int16))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int16) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*2+2], sortvalls[(cur+1)*2:]...)
						copy(sortvalls[(cur+1)*2:(cur+1)*2+2], PackUint16(uint16(value.(int16))))
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return sortvalls
			}
		}
	case []int32:
		if len(sortvalls) == 0 {
			return PackUint32(uint32(value.(int32)))
		}
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := int32(binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4]))
			if value.(int32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*4+4], sortvalls[cur*4:]...)
						copy(sortvalls[cur*4:cur*4+4], PackUint32(uint32(value.(int32))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*4+4], sortvalls[(cur+1)*4:]...)
						copy(sortvalls[(cur+1)*4:(cur+1)*4+4], PackUint32(uint32(value.(int32))))
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return sortvalls
			}
		}
	case []int64:
		if len(sortvalls) == 0 {
			return PackUint64(uint64(value.(int64)))
		}
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := int64(binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8]))
			if value.(int64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*8+8], sortvalls[cur*8:]...)
						copy(sortvalls[cur*8:cur*8+8], PackUint64(uint64(value.(int64))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*8+8], sortvalls[(cur+1)*8:]...)
						copy(sortvalls[(cur+1)*8:(cur+1)*8+8], PackUint64(uint64(value.(int64))))
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return sortvalls
			}
		}
	case []uint8:
		if len(sortvalls) == 0 {
			return []byte{value.(uint8)}
		}
		start := 0
		end := len(sortvalls) - 1
		cur := end / 2
		for true {
			if value.(uint8) < sortvalls[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*1+1], sortvalls[cur*1:]...)
						sortvalls[cur*1] = value.(uint8)
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint8) > sortvalls[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*1+1], sortvalls[(cur+1)*1:]...)
						sortvalls[(cur+1)*1] = value.(uint8)
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return sortvalls
			}
		}
	case []uint16:
		if len(sortvalls) == 0 {
			return PackUint16(value.(uint16))
		}
		start := 0
		end := len(sortvalls)/2 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint16(sortvalls[2*cur : 2*cur+2])
			if value.(uint16) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*2+2], sortvalls[cur*2:]...)
						copy(sortvalls[cur*2:cur*2+2], PackUint16(value.(uint16)))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint16) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*2+2], sortvalls[(cur+1)*2:]...)
						copy(sortvalls[(cur+1)*2:(cur+1)*2+2], PackUint16(value.(uint16)))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return sortvalls
			}
		}
	case []uint32:
		if len(sortvalls) == 0 {
			return PackUint32(value.(uint32))
		}
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4])
			if value.(uint32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*4+4], sortvalls[cur*4:]...)
						copy(sortvalls[cur*4:cur*4+4], PackUint32(value.(uint32)))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*4+4], sortvalls[(cur+1)*4:]...)
						copy(sortvalls[(cur+1)*4:(cur+1)*4+4], PackUint32(value.(uint32)))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return sortvalls
			}
		}
	case []uint64:
		if len(sortvalls) == 0 {
			return PackUint64(value.(uint64))
		}
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8])
			if value.(uint64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*8+8], sortvalls[cur*8:]...)
						copy(sortvalls[cur*8:cur*8+8], PackUint64(value.(uint64)))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*8+8], sortvalls[(cur+1)*8:]...)
						copy(sortvalls[(cur+1)*8:(cur+1)*8+8], PackUint64(value.(uint64)))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return sortvalls
			}
		}
	case []float32:
		if len(sortvalls) == 0 {
			return PackUint32(math.Float32bits(value.(float32)))
		}
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := math.Float32frombits(binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4]))
			if value.(float32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*4+4], sortvalls[cur*4:]...)
						copy(sortvalls[cur*4:cur*4+4], PackUint32(math.Float32bits(value.(float32))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(float32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*4+4], sortvalls[(cur+1)*4:]...)
						copy(sortvalls[(cur+1)*4:(cur+1)*4+4], PackUint32(math.Float32bits(value.(float32))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return sortvalls
			}
		}
	case []float64:
		if len(sortvalls) == 0 {
			return PackUint64(math.Float64bits(value.(float64)))
		}
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := math.Float64frombits(binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8]))
			if value.(float64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						sortvalls = append(sortvalls[:cur*8+8], sortvalls[cur*8:]...)
						copy(sortvalls[cur*8:cur*8+8], PackUint64(math.Float64bits(value.(float64))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(float64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						sortvalls = append(sortvalls[:(cur+1)*8+8], sortvalls[(cur+1)*8:]...)
						copy(sortvalls[(cur+1)*8:(cur+1)*8+8], PackUint64(math.Float64bits(value.(float64))))
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return sortvalls
			}
		}
	default:
		panic("unsupport error")
	}
	return []byte{}
}

func ByteQuickRemove(sortvalls []byte, value interface{}) (outbt []byte) {
	if len(sortvalls) == 0 {
		return []byte{}
	}
	switch value.(type) {
	case int:
		panic("not support error.")
	case int8:
		start := 0
		end := len(sortvalls) - 1
		cur := end / 2
		for true {
			if value.(int8) < int8(sortvalls[cur]) {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int8) > int8(sortvalls[cur]) {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return append(sortvalls[:cur], sortvalls[cur+1:]...)
			}
		}
	case int16:
		start := 0
		end := len(sortvalls)/2 - 1
		cur := end / 2
		for true {
			curval := int16(binary.BigEndian.Uint16(sortvalls[2*cur : 2*cur+2]))
			if value.(int16) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int16) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return append(sortvalls[:2*cur], sortvalls[2*cur+2:]...)
			}
		}
	case int32:
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := int32(binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4]))
			if value.(int32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return append(sortvalls[:4*cur], sortvalls[4*cur+4:]...)
			}
		}
	case int64:
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := int64(binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8]))
			if value.(int64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(int64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return append(sortvalls[:8*cur], sortvalls[8*cur+8:]...)
			}
		}
	case uint8:
		start := 0
		end := len(sortvalls) - 1
		cur := end / 2
		for true {
			if value.(uint8) < sortvalls[cur] {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint8) > sortvalls[cur] {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2 + 1
				}
			} else {
				return append(sortvalls[:1*cur], sortvalls[1*cur+1:]...)
			}
		}
	case uint16:
		start := 0
		end := len(sortvalls)/2 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint16(sortvalls[2*cur : 2*cur+2])
			if value.(uint16) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint16) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return append(sortvalls[:2*cur], sortvalls[2*cur+2:]...)
			}
		}
	case uint32:
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4])
			if value.(uint32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return append(sortvalls[:4*cur], sortvalls[4*cur+4:]...)
			}
		}
	case uint64:
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8])
			if value.(uint64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(uint64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return append(sortvalls[:8*cur], sortvalls[8*cur+8:]...)
			}
		}
	case float32:
		start := 0
		end := len(sortvalls)/4 - 1
		cur := end / 2
		for true {
			curval := math.Float32frombits(binary.BigEndian.Uint32(sortvalls[4*cur : 4*cur+4]))
			if value.(float32) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(float32) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return append(sortvalls[:4*cur], sortvalls[4*cur+4:]...)
			}
		}
	case float64:
		start := 0
		end := len(sortvalls)/8 - 1
		cur := end / 2
		for true {
			curval := math.Float64frombits(binary.BigEndian.Uint64(sortvalls[8*cur : 8*cur+8]))
			if value.(float64) < curval {
				end = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2-1 >= start {
						cur = cur2 - 1
						end = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else if value.(float64) > curval {
				start = cur
				cur2 := int(float64(start+end) / 2)
				if cur2 == cur {
					if cur2+1 <= end {
						cur = cur2 + 1
						start = cur
					} else {
						return sortvalls
					}
				} else {
					cur = cur2
				}
			} else {
				return append(sortvalls[:8*cur], sortvalls[8*cur+8:]...)
			}
		}
	default:
		panic("unsupport error")
	}
	return []byte{}
}

func GetDirSubdirname(dir string) (subdir []string) {
	dir = StdDir(dir)
	d, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return nil
	}
	for _, name := range names {
		info, err := os.Stat(dir + name)
		if err != nil {
			return subdir
		} else if info.IsDir() {
			subdir = append(subdir, name)
		}
	}
	return subdir
}

func GetDirSubfilename(dir string) (subfile []string) {
	dir = StdDir(dir)
	d, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return nil
	}
	for _, name := range names {
		info, err := os.Stat(dir + name)
		if err != nil {
			return subfile
		} else if !info.IsDir() {
			subfile = append(subfile, name)
		}
	}
	return subfile
}

func PackUint32String(num string) string {
	u32, u32e := strconv.ParseUint(num, 10, 32)
	if u32e != nil {
		return ""
	}
	return string(PackUint32(uint32(u32)))
}

func PackUint64String(num string) string {
	u64, u64e := strconv.ParseUint(num, 10, 64)
	if u64e != nil {
		return ""
	}
	return string(PackUint64(u64))
}

func UnpackUint32String(btstr string) string {
	return strconv.FormatUint(uint64(binary.BigEndian.Uint32([]byte(btstr))), 10)
}

func UnpackUint64String(btstr string) string {
	return strconv.FormatUint(uint64(binary.BigEndian.Uint64([]byte(btstr))), 10)
}

func Create(path string) (ff *os.File, err error) {
	ff, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	return ff, err
}

func OpenFile(path string, flag int, perm os.FileMode) (ff *os.File, err error) {
	ff, err = os.OpenFile(path, os.O_RDWR, 0666)
	if err == nil {
		ff.Seek(0, os.SEEK_END)
		return ff, err
	} else {
		return Create(path)
	}
}

func YestodayUTCSec() int64 {
	nowsec := time.Now().Unix()
	return nowsec - nowsec%(24*3600)
}

func YestodayLocalSec() int64 {
	nowsec := time.Now().Unix()
	return nowsec - nowsec%(24*3600)
}

func TodayUtcSec() int64 {
	nowsec := time.Now().Unix()
	return nowsec % (24 * 3600)
}

func TodayLocalSec() int64 {
	nowsec := time.Now().Local().Unix()
	return nowsec % (24 * 3600)
}

func IsoTimeToSec(isotime string) int64 {
	strls := regexp.MustCompile("(?ism)(\\d{1,2}):(\\d{1,2}):(\\d{1,2})$").FindAllStringSubmatch(isotime, 1)
	hour, houre := strconv.ParseInt(strls[0][1], 10, 64)
	minute, minutee := strconv.ParseInt(strls[0][2], 10, 64)
	second, seconde := strconv.ParseInt(strls[0][3], 10, 64)
	if houre == nil && minutee == nil && seconde == nil {
		return hour*3600 + minute*60 + second
	}
	return -1
}

func IsoDateTimeToSec(isodatetime string) int64 {
	if strings.Index(isodatetime, " ") != -1 {
		t, te := time.Parse("2006-01-02 15:04:05", isodatetime)
		if te == nil {
			t.Unix()
		}
	} else if strings.Index(isodatetime, "T") != -1 {
		t, te := time.Parse("2006-01-02T15:04:05", isodatetime)
		if te == nil {
			t.Unix()
		}
	}
	panic("formt error")
	return -1
}

func IsoDateTimeToTime(isodatetime string) *time.Time {
	if strings.Index(isodatetime, " ") != -1 {
		t, te := time.Parse("2006-01-02 15:04:05", isodatetime)
		if te == nil {
			t2 := &time.Time{}
			*t2 = t
			return t2
		}
	} else if strings.Index(isodatetime, "T") != -1 {
		t, te := time.Parse("2006-01-02T15:04:05", isodatetime)
		if te == nil {
			t2 := &time.Time{}
			*t2 = t
			return t2
		}
	}
	return nil
}

//https://raw.githubusercontent.com/eggert/tz/master/zone1970.tab
var zone1970 string = `# tzdb timezone descriptions
#
# This file is in the public domain.
#
# From Paul Eggert (2018-06-27):
# This file contains a table where each row stands for a timezone where
# civil timestamps have agreed since 1970.  Columns are separated by
# a single tab.  Lines beginning with '#' are comments.  All text uses
# UTF-8 encoding.  The columns of the table are as follows:
#
# 1.  The countries that overlap the timezone, as a comma-separated list
#     of ISO 3166 2-character country codes.  See the file 'iso3166.tab'.
# 2.  Latitude and longitude of the timezone's principal location
#     in ISO 6709 sign-degrees-minutes-seconds format,
#     either ±DDMM±DDDMM or ±DDMMSS±DDDMMSS,
#     first latitude (+ is north), then longitude (+ is east).
# 3.  Timezone name used in value of TZ environment variable.
#     Please see the theory.html file for how these names are chosen.
#     If multiple timezones overlap a country, each has a row in the
#     table, with each column 1 containing the country code.
# 4.  Comments; present if and only if a country has multiple timezones.
#
# If a timezone covers multiple countries, the most-populous city is used,
# and that country is listed first in column 1; any other countries
# are listed alphabetically by country code.  The table is sorted
# first by country code, then (if possible) by an order within the
# country that (1) makes some geographical sense, and (2) puts the
# most populous timezones first, where that does not contradict (1).
#
# This table is intended as an aid for users, to help them select timezones
# appropriate for their practical needs.  It is not intended to take or
# endorse any position on legal or territorial claims.
#
#country-
#codes	coordinates	TZ	comments
AD	+4230+00131	Europe/Andorra
AE,OM	+2518+05518	Asia/Dubai
AF	+3431+06912	Asia/Kabul
AL	+4120+01950	Europe/Tirane
AM	+4011+04430	Asia/Yerevan
AQ	-6617+11031	Antarctica/Casey	Casey
AQ	-6835+07758	Antarctica/Davis	Davis
AQ	-6640+14001	Antarctica/DumontDUrville	Dumont-d'Urville
AQ	-6736+06253	Antarctica/Mawson	Mawson
AQ	-6448-06406	Antarctica/Palmer	Palmer
AQ	-6734-06808	Antarctica/Rothera	Rothera
AQ	-690022+0393524	Antarctica/Syowa	Syowa
AQ	-720041+0023206	Antarctica/Troll	Troll
AQ	-7824+10654	Antarctica/Vostok	Vostok
AR	-3436-05827	America/Argentina/Buenos_Aires	Buenos Aires (BA, CF)
AR	-3124-06411	America/Argentina/Cordoba	Argentina (most areas: CB, CC, CN, ER, FM, MN, SE, SF)
AR	-2447-06525	America/Argentina/Salta	Salta (SA, LP, NQ, RN)
AR	-2411-06518	America/Argentina/Jujuy	Jujuy (JY)
AR	-2649-06513	America/Argentina/Tucuman	Tucumán (TM)
AR	-2828-06547	America/Argentina/Catamarca	Catamarca (CT); Chubut (CH)
AR	-2926-06651	America/Argentina/La_Rioja	La Rioja (LR)
AR	-3132-06831	America/Argentina/San_Juan	San Juan (SJ)
AR	-3253-06849	America/Argentina/Mendoza	Mendoza (MZ)
AR	-3319-06621	America/Argentina/San_Luis	San Luis (SL)
AR	-5138-06913	America/Argentina/Rio_Gallegos	Santa Cruz (SC)
AR	-5448-06818	America/Argentina/Ushuaia	Tierra del Fuego (TF)
AS,UM	-1416-17042	Pacific/Pago_Pago	Samoa, Midway
AT	+4813+01620	Europe/Vienna
AU	-3133+15905	Australia/Lord_Howe	Lord Howe Island
AU	-5430+15857	Antarctica/Macquarie	Macquarie Island
AU	-4253+14719	Australia/Hobart	Tasmania (most areas)
AU	-3956+14352	Australia/Currie	Tasmania (King Island)
AU	-3749+14458	Australia/Melbourne	Victoria
AU	-3352+15113	Australia/Sydney	New South Wales (most areas)
AU	-3157+14127	Australia/Broken_Hill	New South Wales (Yancowinna)
AU	-2728+15302	Australia/Brisbane	Queensland (most areas)
AU	-2016+14900	Australia/Lindeman	Queensland (Whitsunday Islands)
AU	-3455+13835	Australia/Adelaide	South Australia
AU	-1228+13050	Australia/Darwin	Northern Territory
AU	-3157+11551	Australia/Perth	Western Australia (most areas)
AU	-3143+12852	Australia/Eucla	Western Australia (Eucla)
AZ	+4023+04951	Asia/Baku
BB	+1306-05937	America/Barbados
BD	+2343+09025	Asia/Dhaka
BE	+5050+00420	Europe/Brussels
BG	+4241+02319	Europe/Sofia
BM	+3217-06446	Atlantic/Bermuda
BN	+0456+11455	Asia/Brunei
BO	-1630-06809	America/La_Paz
BR	-0351-03225	America/Noronha	Atlantic islands
BR	-0127-04829	America/Belem	Pará (east); Amapá
BR	-0343-03830	America/Fortaleza	Brazil (northeast: MA, PI, CE, RN, PB)
BR	-0803-03454	America/Recife	Pernambuco
BR	-0712-04812	America/Araguaina	Tocantins
BR	-0940-03543	America/Maceio	Alagoas, Sergipe
BR	-1259-03831	America/Bahia	Bahia
BR	-2332-04637	America/Sao_Paulo	Brazil (southeast: GO, DF, MG, ES, RJ, SP, PR, SC, RS)
BR	-2027-05437	America/Campo_Grande	Mato Grosso do Sul
BR	-1535-05605	America/Cuiaba	Mato Grosso
BR	-0226-05452	America/Santarem	Pará (west)
BR	-0846-06354	America/Porto_Velho	Rondônia
BR	+0249-06040	America/Boa_Vista	Roraima
BR	-0308-06001	America/Manaus	Amazonas (east)
BR	-0640-06952	America/Eirunepe	Amazonas (west)
BR	-0958-06748	America/Rio_Branco	Acre
BS	+2505-07721	America/Nassau
BT	+2728+08939	Asia/Thimphu
BY	+5354+02734	Europe/Minsk
BZ	+1730-08812	America/Belize
CA	+4734-05243	America/St_Johns	Newfoundland; Labrador (southeast)
CA	+4439-06336	America/Halifax	Atlantic - NS (most areas); PE
CA	+4612-05957	America/Glace_Bay	Atlantic - NS (Cape Breton)
CA	+4606-06447	America/Moncton	Atlantic - New Brunswick
CA	+5320-06025	America/Goose_Bay	Atlantic - Labrador (most areas)
CA	+5125-05707	America/Blanc-Sablon	AST - QC (Lower North Shore)
CA	+4339-07923	America/Toronto	Eastern - ON, QC (most areas)
CA	+4901-08816	America/Nipigon	Eastern - ON, QC (no DST 1967-73)
CA	+4823-08915	America/Thunder_Bay	Eastern - ON (Thunder Bay)
CA	+6344-06828	America/Iqaluit	Eastern - NU (most east areas)
CA	+6608-06544	America/Pangnirtung	Eastern - NU (Pangnirtung)
CA	+484531-0913718	America/Atikokan	EST - ON (Atikokan); NU (Coral H)
CA	+4953-09709	America/Winnipeg	Central - ON (west); Manitoba
CA	+4843-09434	America/Rainy_River	Central - ON (Rainy R, Ft Frances)
CA	+744144-0944945	America/Resolute	Central - NU (Resolute)
CA	+624900-0920459	America/Rankin_Inlet	Central - NU (central)
CA	+5024-10439	America/Regina	CST - SK (most areas)
CA	+5017-10750	America/Swift_Current	CST - SK (midwest)
CA	+5333-11328	America/Edmonton	Mountain - AB; BC (E); SK (W)
CA	+690650-1050310	America/Cambridge_Bay	Mountain - NU (west)
CA	+6227-11421	America/Yellowknife	Mountain - NT (central)
CA	+682059-1334300	America/Inuvik	Mountain - NT (west)
CA	+4906-11631	America/Creston	MST - BC (Creston)
CA	+5946-12014	America/Dawson_Creek	MST - BC (Dawson Cr, Ft St John)
CA	+5848-12242	America/Fort_Nelson	MST - BC (Ft Nelson)
CA	+4916-12307	America/Vancouver	Pacific - BC (most areas)
CA	+6043-13503	America/Whitehorse	Pacific - Yukon (south)
CA	+6404-13925	America/Dawson	Pacific - Yukon (north)
CC	-1210+09655	Indian/Cocos
CH,DE,LI	+4723+00832	Europe/Zurich	Swiss time
CI,BF,GM,GN,ML,MR,SH,SL,SN,TG	+0519-00402	Africa/Abidjan
CK	-2114-15946	Pacific/Rarotonga
CL	-3327-07040	America/Santiago	Chile (most areas)
CL	-5309-07055	America/Punta_Arenas	Region of Magallanes
CL	-2709-10926	Pacific/Easter	Easter Island
CN	+3114+12128	Asia/Shanghai	Beijing Time
CN	+4348+08735	Asia/Urumqi	Xinjiang Time
CO	+0436-07405	America/Bogota
CR	+0956-08405	America/Costa_Rica
CU	+2308-08222	America/Havana
CV	+1455-02331	Atlantic/Cape_Verde
CW,AW,BQ,SX	+1211-06900	America/Curacao
CX	-1025+10543	Indian/Christmas
CY	+3510+03322	Asia/Nicosia	Cyprus (most areas)
CY	+3507+03357	Asia/Famagusta	Northern Cyprus
CZ,SK	+5005+01426	Europe/Prague
DE	+5230+01322	Europe/Berlin	Germany (most areas)
DK	+5540+01235	Europe/Copenhagen
DO	+1828-06954	America/Santo_Domingo
DZ	+3647+00303	Africa/Algiers
EC	-0210-07950	America/Guayaquil	Ecuador (mainland)
EC	-0054-08936	Pacific/Galapagos	Galápagos Islands
EE	+5925+02445	Europe/Tallinn
EG	+3003+03115	Africa/Cairo
EH	+2709-01312	Africa/El_Aaiun
ES	+4024-00341	Europe/Madrid	Spain (mainland)
ES	+3553-00519	Africa/Ceuta	Ceuta, Melilla
ES	+2806-01524	Atlantic/Canary	Canary Islands
FI,AX	+6010+02458	Europe/Helsinki
FJ	-1808+17825	Pacific/Fiji
FK	-5142-05751	Atlantic/Stanley
FM	+0725+15147	Pacific/Chuuk	Chuuk/Truk, Yap
FM	+0658+15813	Pacific/Pohnpei	Pohnpei/Ponape
FM	+0519+16259	Pacific/Kosrae	Kosrae
FO	+6201-00646	Atlantic/Faroe
FR	+4852+00220	Europe/Paris
GB,GG,IM,JE	+513030-0000731	Europe/London
GE	+4143+04449	Asia/Tbilisi
GF	+0456-05220	America/Cayenne
GH	+0533-00013	Africa/Accra
GI	+3608-00521	Europe/Gibraltar
GL	+6411-05144	America/Godthab	Greenland (most areas)
GL	+7646-01840	America/Danmarkshavn	National Park (east coast)
GL	+7029-02158	America/Scoresbysund	Scoresbysund/Ittoqqortoormiit
GL	+7634-06847	America/Thule	Thule/Pituffik
GR	+3758+02343	Europe/Athens
GS	-5416-03632	Atlantic/South_Georgia
GT	+1438-09031	America/Guatemala
GU,MP	+1328+14445	Pacific/Guam
GW	+1151-01535	Africa/Bissau
GY	+0648-05810	America/Guyana
HK	+2217+11409	Asia/Hong_Kong
HN	+1406-08713	America/Tegucigalpa
HT	+1832-07220	America/Port-au-Prince
HU	+4730+01905	Europe/Budapest
ID	-0610+10648	Asia/Jakarta	Java, Sumatra
ID	-0002+10920	Asia/Pontianak	Borneo (west, central)
ID	-0507+11924	Asia/Makassar	Borneo (east, south); Sulawesi/Celebes, Bali, Nusa Tengarra; Timor (west)
ID	-0232+14042	Asia/Jayapura	New Guinea (West Papua / Irian Jaya); Malukus/Moluccas
IE	+5320-00615	Europe/Dublin
IL	+314650+0351326	Asia/Jerusalem
IN	+2232+08822	Asia/Kolkata
IO	-0720+07225	Indian/Chagos
IQ	+3321+04425	Asia/Baghdad
IR	+3540+05126	Asia/Tehran
IS	+6409-02151	Atlantic/Reykjavik
IT,SM,VA	+4154+01229	Europe/Rome
JM	+175805-0764736	America/Jamaica
JO	+3157+03556	Asia/Amman
JP	+353916+1394441	Asia/Tokyo
KE,DJ,ER,ET,KM,MG,SO,TZ,UG,YT	-0117+03649	Africa/Nairobi
KG	+4254+07436	Asia/Bishkek
KI	+0125+17300	Pacific/Tarawa	Gilbert Islands
KI	-0308-17105	Pacific/Enderbury	Phoenix Islands
KI	+0152-15720	Pacific/Kiritimati	Line Islands
KP	+3901+12545	Asia/Pyongyang
KR	+3733+12658	Asia/Seoul
KZ	+4315+07657	Asia/Almaty	Kazakhstan (most areas)
KZ	+4448+06528	Asia/Qyzylorda	Qyzylorda/Kyzylorda/Kzyl-Orda
KZ	+5312+06337	Asia/Qostanay	Qostanay/Kostanay/Kustanay
KZ	+5017+05710	Asia/Aqtobe	Aqtöbe/Aktobe
KZ	+4431+05016	Asia/Aqtau	Mangghystaū/Mankistau
KZ	+4707+05156	Asia/Atyrau	Atyraū/Atirau/Gur'yev
KZ	+5113+05121	Asia/Oral	West Kazakhstan
LB	+3353+03530	Asia/Beirut
LK	+0656+07951	Asia/Colombo
LR	+0618-01047	Africa/Monrovia
LT	+5441+02519	Europe/Vilnius
LU	+4936+00609	Europe/Luxembourg
LV	+5657+02406	Europe/Riga
LY	+3254+01311	Africa/Tripoli
MA	+3339-00735	Africa/Casablanca
MC	+4342+00723	Europe/Monaco
MD	+4700+02850	Europe/Chisinau
MH	+0709+17112	Pacific/Majuro	Marshall Islands (most areas)
MH	+0905+16720	Pacific/Kwajalein	Kwajalein
MM	+1647+09610	Asia/Yangon
MN	+4755+10653	Asia/Ulaanbaatar	Mongolia (most areas)
MN	+4801+09139	Asia/Hovd	Bayan-Ölgii, Govi-Altai, Hovd, Uvs, Zavkhan
MN	+4804+11430	Asia/Choibalsan	Dornod, Sükhbaatar
MO	+221150+1133230	Asia/Macau
MQ	+1436-06105	America/Martinique
MT	+3554+01431	Europe/Malta
MU	-2010+05730	Indian/Mauritius
MV	+0410+07330	Indian/Maldives
MX	+1924-09909	America/Mexico_City	Central Time
MX	+2105-08646	America/Cancun	Eastern Standard Time - Quintana Roo
MX	+2058-08937	America/Merida	Central Time - Campeche, Yucatán
MX	+2540-10019	America/Monterrey	Central Time - Durango; Coahuila, Nuevo León, Tamaulipas (most areas)
MX	+2550-09730	America/Matamoros	Central Time US - Coahuila, Nuevo León, Tamaulipas (US border)
MX	+2313-10625	America/Mazatlan	Mountain Time - Baja California Sur, Nayarit, Sinaloa
MX	+2838-10605	America/Chihuahua	Mountain Time - Chihuahua (most areas)
MX	+2934-10425	America/Ojinaga	Mountain Time US - Chihuahua (US border)
MX	+2904-11058	America/Hermosillo	Mountain Standard Time - Sonora
MX	+3232-11701	America/Tijuana	Pacific Time US - Baja California
MX	+2048-10515	America/Bahia_Banderas	Central Time - Bahía de Banderas
MY	+0310+10142	Asia/Kuala_Lumpur	Malaysia (peninsula)
MY	+0133+11020	Asia/Kuching	Sabah, Sarawak
MZ,BI,BW,CD,MW,RW,ZM,ZW	-2558+03235	Africa/Maputo	Central Africa Time
NA	-2234+01706	Africa/Windhoek
NC	-2216+16627	Pacific/Noumea
NF	-2903+16758	Pacific/Norfolk
NG,AO,BJ,CD,CF,CG,CM,GA,GQ,NE	+0627+00324	Africa/Lagos	West Africa Time
NI	+1209-08617	America/Managua
NL	+5222+00454	Europe/Amsterdam
NO,SJ	+5955+01045	Europe/Oslo
NP	+2743+08519	Asia/Kathmandu
NR	-0031+16655	Pacific/Nauru
NU	-1901-16955	Pacific/Niue
NZ,AQ	-3652+17446	Pacific/Auckland	New Zealand time
NZ	-4357-17633	Pacific/Chatham	Chatham Islands
PA,KY	+0858-07932	America/Panama
PE	-1203-07703	America/Lima
PF	-1732-14934	Pacific/Tahiti	Society Islands
PF	-0900-13930	Pacific/Marquesas	Marquesas Islands
PF	-2308-13457	Pacific/Gambier	Gambier Islands
PG	-0930+14710	Pacific/Port_Moresby	Papua New Guinea (most areas)
PG	-0613+15534	Pacific/Bougainville	Bougainville
PH	+1435+12100	Asia/Manila
PK	+2452+06703	Asia/Karachi
PL	+5215+02100	Europe/Warsaw
PM	+4703-05620	America/Miquelon
PN	-2504-13005	Pacific/Pitcairn
PR	+182806-0660622	America/Puerto_Rico
PS	+3130+03428	Asia/Gaza	Gaza Strip
PS	+313200+0350542	Asia/Hebron	West Bank
PT	+3843-00908	Europe/Lisbon	Portugal (mainland)
PT	+3238-01654	Atlantic/Madeira	Madeira Islands
PT	+3744-02540	Atlantic/Azores	Azores
PW	+0720+13429	Pacific/Palau
PY	-2516-05740	America/Asuncion
QA,BH	+2517+05132	Asia/Qatar
RE,TF	-2052+05528	Indian/Reunion	Réunion, Crozet, Scattered Islands
RO	+4426+02606	Europe/Bucharest
RS,BA,HR,ME,MK,SI	+4450+02030	Europe/Belgrade
RU	+5443+02030	Europe/Kaliningrad	MSK-01 - Kaliningrad
RU	+554521+0373704	Europe/Moscow	MSK+00 - Moscow area
RU	+4457+03406	Europe/Simferopol	MSK+00 - Crimea
RU	+5836+04939	Europe/Kirov	MSK+00 - Kirov
RU	+4621+04803	Europe/Astrakhan	MSK+01 - Astrakhan
RU	+4844+04425	Europe/Volgograd	MSK+01 - Volgograd
RU	+5134+04602	Europe/Saratov	MSK+01 - Saratov
RU	+5420+04824	Europe/Ulyanovsk	MSK+01 - Ulyanovsk
RU	+5312+05009	Europe/Samara	MSK+01 - Samara, Udmurtia
RU	+5651+06036	Asia/Yekaterinburg	MSK+02 - Urals
RU	+5500+07324	Asia/Omsk	MSK+03 - Omsk
RU	+5502+08255	Asia/Novosibirsk	MSK+04 - Novosibirsk
RU	+5322+08345	Asia/Barnaul	MSK+04 - Altai
RU	+5630+08458	Asia/Tomsk	MSK+04 - Tomsk
RU	+5345+08707	Asia/Novokuznetsk	MSK+04 - Kemerovo
RU	+5601+09250	Asia/Krasnoyarsk	MSK+04 - Krasnoyarsk area
RU	+5216+10420	Asia/Irkutsk	MSK+05 - Irkutsk, Buryatia
RU	+5203+11328	Asia/Chita	MSK+06 - Zabaykalsky
RU	+6200+12940	Asia/Yakutsk	MSK+06 - Lena River
RU	+623923+1353314	Asia/Khandyga	MSK+06 - Tomponsky, Ust-Maysky
RU	+4310+13156	Asia/Vladivostok	MSK+07 - Amur River
RU	+643337+1431336	Asia/Ust-Nera	MSK+07 - Oymyakonsky
RU	+5934+15048	Asia/Magadan	MSK+08 - Magadan
RU	+4658+14242	Asia/Sakhalin	MSK+08 - Sakhalin Island
RU	+6728+15343	Asia/Srednekolymsk	MSK+08 - Sakha (E); North Kuril Is
RU	+5301+15839	Asia/Kamchatka	MSK+09 - Kamchatka
RU	+6445+17729	Asia/Anadyr	MSK+09 - Bering Sea
SA,KW,YE	+2438+04643	Asia/Riyadh
SB	-0932+16012	Pacific/Guadalcanal
SC	-0440+05528	Indian/Mahe
SD	+1536+03232	Africa/Khartoum
SE	+5920+01803	Europe/Stockholm
SG	+0117+10351	Asia/Singapore
SR	+0550-05510	America/Paramaribo
SS	+0451+03137	Africa/Juba
ST	+0020+00644	Africa/Sao_Tome
SV	+1342-08912	America/El_Salvador
SY	+3330+03618	Asia/Damascus
TC	+2128-07108	America/Grand_Turk
TD	+1207+01503	Africa/Ndjamena
TF	-492110+0701303	Indian/Kerguelen	Kerguelen, St Paul Island, Amsterdam Island
TH,KH,LA,VN	+1345+10031	Asia/Bangkok	Indochina (most areas)
TJ	+3835+06848	Asia/Dushanbe
TK	-0922-17114	Pacific/Fakaofo
TL	-0833+12535	Asia/Dili
TM	+3757+05823	Asia/Ashgabat
TN	+3648+01011	Africa/Tunis
TO	-2110-17510	Pacific/Tongatapu
TR	+4101+02858	Europe/Istanbul
TT,AG,AI,BL,DM,GD,GP,KN,LC,MF,MS,VC,VG,VI	+1039-06131	America/Port_of_Spain
TV	-0831+17913	Pacific/Funafuti
TW	+2503+12130	Asia/Taipei
UA	+5026+03031	Europe/Kiev	Ukraine (most areas)
UA	+4837+02218	Europe/Uzhgorod	Ruthenia
UA	+4750+03510	Europe/Zaporozhye	Zaporozh'ye/Zaporizhia; Lugansk/Luhansk (east)
UM	+1917+16637	Pacific/Wake	Wake Island
US	+404251-0740023	America/New_York	Eastern (most areas)
US	+421953-0830245	America/Detroit	Eastern - MI (most areas)
US	+381515-0854534	America/Kentucky/Louisville	Eastern - KY (Louisville area)
US	+364947-0845057	America/Kentucky/Monticello	Eastern - KY (Wayne)
US	+394606-0860929	America/Indiana/Indianapolis	Eastern - IN (most areas)
US	+384038-0873143	America/Indiana/Vincennes	Eastern - IN (Da, Du, K, Mn)
US	+410305-0863611	America/Indiana/Winamac	Eastern - IN (Pulaski)
US	+382232-0862041	America/Indiana/Marengo	Eastern - IN (Crawford)
US	+382931-0871643	America/Indiana/Petersburg	Eastern - IN (Pike)
US	+384452-0850402	America/Indiana/Vevay	Eastern - IN (Switzerland)
US	+415100-0873900	America/Chicago	Central (most areas)
US	+375711-0864541	America/Indiana/Tell_City	Central - IN (Perry)
US	+411745-0863730	America/Indiana/Knox	Central - IN (Starke)
US	+450628-0873651	America/Menominee	Central - MI (Wisconsin border)
US	+470659-1011757	America/North_Dakota/Center	Central - ND (Oliver)
US	+465042-1012439	America/North_Dakota/New_Salem	Central - ND (Morton rural)
US	+471551-1014640	America/North_Dakota/Beulah	Central - ND (Mercer)
US	+394421-1045903	America/Denver	Mountain (most areas)
US	+433649-1161209	America/Boise	Mountain - ID (south); OR (east)
US	+332654-1120424	America/Phoenix	MST - Arizona (except Navajo)
US	+340308-1181434	America/Los_Angeles	Pacific
US	+611305-1495401	America/Anchorage	Alaska (most areas)
US	+581807-1342511	America/Juneau	Alaska - Juneau area
US	+571035-1351807	America/Sitka	Alaska - Sitka area
US	+550737-1313435	America/Metlakatla	Alaska - Annette Island
US	+593249-1394338	America/Yakutat	Alaska - Yakutat
US	+643004-1652423	America/Nome	Alaska (west)
US	+515248-1763929	America/Adak	Aleutian Islands
US,UM	+211825-1575130	Pacific/Honolulu	Hawaii
UY	-345433-0561245	America/Montevideo
UZ	+3940+06648	Asia/Samarkand	Uzbekistan (west)
UZ	+4120+06918	Asia/Tashkent	Uzbekistan (east)
VE	+1030-06656	America/Caracas
VN	+1045+10640	Asia/Ho_Chi_Minh	Vietnam (south)
VU	-1740+16825	Pacific/Efate
WF	-1318-17610	Pacific/Wallis
WS	-1350-17144	Pacific/Apia
ZA,LS,SZ	-2615+02800	Africa/Johannesburg
`

func CountryZones() map[string][]string {
	countries := make(map[string][]string)
	bf := bufio.NewReader(strings.NewReader(zone1970))
	for true {
		line, _, linee := bf.ReadLine()
		if linee != nil {
			break
		}
		line2 := string(bytes.Trim(line, "\r\n"))
		if strings.HasPrefix(line2, "#") {
			continue
		}
		n := 3
		fields := strings.SplitN(line2, "\t", n+1)
		if len(fields) < n {
			continue
		}
		zone := fields[2]
		for _, country := range strings.Split(fields[0], ",") {
			country = strings.ToUpper(country)
			zones := countries[country]
			zones = append(zones, zone)
			countries[country] = zones
		}
	}
	return countries
}

func TimeZoneSec(timezone string) int64 {
	if l, err := time.LoadLocation(timezone); err == nil {
		timeStr := "2006-01-02 15:04:05"
		lt, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, l)
		l2, _ := time.LoadLocation("UTC")
		lt2, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, l2)
		return lt.Unix() - lt2.Unix()
	}
	return -1
}

func GetCurrentZoneSec() int64 {
	t := time.Now()
	_, zonesec := t.Zone()
	return int64(zonesec)
}

func TimeFromInt64(i int64) time.Time {
	return time.Unix(i, 0)
}

func IntToStr(v int) string {
	return strconv.FormatInt(int64(v), 10)
}

func Int32ToStr(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

func Int64ToStr(v int64) string {
	return strconv.FormatInt(v, 10)
}

func Uint32ToStr(v uint32) string {
	return strconv.FormatUint(uint64(v), 10)
}

func Uint64ToStr(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func Float32ToStr(v float32) string {
	return strconv.FormatFloat(float64(v), 'g', 16, 32)
}
func Float64ToStr(v float64) string {
	return strconv.FormatFloat(v, 'g', 16, 64)
}

func Float32RoundToStr(v float32, roundn int) string {
	return fmt.Sprintf("%."+strconv.Itoa(roundn)+"g", v)
}
func Float64RoundToStr(v float64, roundn int) string {
	return fmt.Sprintf("%."+strconv.Itoa(roundn)+"g", v)
}

func FixLenWithFillRightStr(val string, fixlen int, fch string) string {
	for i := len(val); i < fixlen; i++ {
		val = val + fch
	}
	return val
}
func FixLenWithFillLeftStr(val string, fixlen int, fch string) string {
	for i := len(val); i < fixlen; i++ {
		val = fch + val
	}
	return val
}

func FixLenWithFillRight(val string, fixlen int, fch byte) string {
	for i := len(val); i < fixlen; i++ {
		val = val + string([]byte{fch})
	}
	return val
}
func FixLenWithFillLeft(val string, fixlen int, fch byte) string {
	for i := len(val); i < fixlen; i++ {
		val = string([]byte{fch}) + val
	}
	return val
}

func IntFromStr(str string) int {
	i64, i64e := strconv.ParseInt(str, 10, 64)
	if i64e != nil {
		fmt.Println("IntFromStr", str)
		panic("error")
	}
	return int(i64)
}

func Int32FromStr(str string) int32 {
	i64, i64e := strconv.ParseInt(str, 10, 32)
	if i64e != nil {
		fmt.Println("Int32FromStr", str)
		panic("error")
	}
	return int32(i64)
}

func Int64FromStr(str string) int64 {
	i64, i64e := strconv.ParseUint(str, 10, 64)
	if i64e != nil {
		fmt.Println("Int64FromStr", str)
		panic("error")
	}
	return int64(i64)
}

func Uint32FromStr(str string) uint32 {
	i64, i64e := strconv.ParseUint(str, 10, 32)
	if i64e != nil {
		fmt.Println("Uint32FromStr", str)
		panic("error")
	}
	return uint32(i64)
}

func Uint64FromStr(str string) uint64 {
	i64, i64e := strconv.ParseUint(str, 10, 64)
	if i64e != nil {
		fmt.Println("Uint64FromStr", str)
		panic("error")
	}
	return uint64(i64)
}

func Float32FromStr(str string) float32 {
	i64, i64e := strconv.ParseFloat(str, 32)
	if i64e != nil {
		fmt.Println("Float32FromStr", str)
		panic("error")
	}
	return float32(i64)
}
func Float64FromStr(str string) float64 {
	i64, i64e := strconv.ParseFloat(str, 64)
	if i64e != nil {
		fmt.Println("Float64FromStr", str)
		panic("error")
	}
	return float64(i64)
}

func HttpTimeToTime(timestr, timezone string) time.Time {
	timesegs := strings.Split(timestr, " ")
	houmin := strings.Split(timesegs[3], ":")
	sec := 0
	if len(houmin) == 3 {
		sec = IntFromStr(houmin[2])
	}
	l, _ := time.LoadLocation(timezone)
	return time.Date(IntFromStr(timesegs[2]), time.Month(AbbreviateMonthToNum(timesegs[1])), IntFromStr(timesegs[0]), IntFromStr(houmin[0]), IntFromStr(houmin[1]), sec, 0, l)
}

func Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

func Uint32ListToBytes(u32list []uint32) (result []byte) {
	result = make([]byte, 4*len(u32list))
	for i := 0; i < len(u32list); i++ {
		binary.BigEndian.PutUint32(result[i*4:i*4+4], u32list[i])
	}
	return result
}

func Uint64ListToBytes(u64list []uint64) (result []byte) {
	result = make([]byte, 4*len(u64list))
	for i := 0; i < len(u64list); i++ {
		binary.BigEndian.PutUint64(result[i*8:i*8+8], u64list[i])
	}
	return result
}

func IsNumber(word []byte) bool {
	numberre := regexp.MustCompile("(?ism)^[+-]?[0-9]+(\\.[0-9]+)?([Ee][+-]?[0-9]+(\\.[0-9]+)?)?$")
	return numberre.Match([]byte(word))
}

func IsIPV4(word []byte) bool {
	numberre := regexp.MustCompile("(?ism)^([0-9]{1,3})[.]([0-9]{1,3})[.]([0-9]{1,3})[.]([0-9]{1,3})$")
	indls := numberre.FindAllSubmatchIndex([]byte(word), -1)
	if len(indls) == 1 {
		for i := 1; i < len(indls[0])/2; i += 1 {
			num, _ := strconv.ParseInt(string(word[indls[0][i*2]:indls[0][i*2+1]]), 10, 64)
			if !(num >= 0 && num <= 255) {
				return false
			}
		}
		return true
	}
	return false
}

func IsIsoDateTime(word []byte) bool {
	numberre := regexp.MustCompile("(?ism)^([0-9]{4})-([0-9]{2})-([0-9]{2})T([0-9]{2}):([0-9]{2}):([0-9]{2})$")
	indls := numberre.FindAllSubmatchIndex([]byte(word), -1)
	if len(indls) == 1 {
		num, _ := strconv.ParseInt(string(word[indls[0][1*2]:indls[0][1*2+1]]), 10, 64)
		if !(num >= 0 && num <= 9999) {
			return false
		}
		month, _ := strconv.ParseInt(string(word[indls[0][2*2]:indls[0][2*2+1]]), 10, 64)
		if !(month >= 1 && month <= 12) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][3*2]:indls[0][3*2+1]]), 10, 64)
		if !((month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12) && num >= 1 && num <= 31 || (month == 4 || month == 6 || month == 9 || month == 11) && num >= 1 && num <= 30 || (month == 2) && num >= 1 && num <= 29) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][4*2]:indls[0][4*2+1]]), 10, 64)
		if !(num >= 0 && num <= 23) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][5*2]:indls[0][5*2+1]]), 10, 64)
		if !(num >= 0 && num <= 59) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][6*2]:indls[0][6*2+1]]), 10, 64)
		if !(num >= 0 && num <= 59) {
			return false
		}
		return true
	}
	return false
}

func IsIsoDate(word []byte) bool {
	numberre := regexp.MustCompile("(?ism)^([0-9]{4})-([0-9]{2})-([0-9]{2})$")
	indls := numberre.FindAllSubmatchIndex([]byte(word), -1)
	if len(indls) == 1 {
		num, _ := strconv.ParseInt(string(word[indls[0][1*2]:indls[0][1*2+1]]), 10, 64)
		if !(num >= 0 && num <= 9999) {
			return false
		}
		month, _ := strconv.ParseInt(string(word[indls[0][2*2]:indls[0][2*2+1]]), 10, 64)
		if !(month >= 1 && month <= 12) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][3*2]:indls[0][3*2+1]]), 10, 64)
		if !((month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12) && num >= 1 && num <= 31 || (month == 4 || month == 6 || month == 9 || month == 11) && num >= 1 && num <= 30 || (month == 2) && num >= 1 && num <= 29) {
			return false
		}
		return true
	}
	return false
}

func IsIsoTime(word []byte) bool {
	numberre := regexp.MustCompile("(?ism)^([0-9]{2}):([0-9]{2}):([0-9]{2})$")
	indls := numberre.FindAllSubmatchIndex([]byte(word), -1)
	if len(indls) == 1 {
		num, _ := strconv.ParseInt(string(word[indls[0][1*2]:indls[0][1*2+1]]), 10, 64)
		if !(num >= 0 && num <= 23) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][2*2]:indls[0][2*2+1]]), 10, 64)
		if !(num >= 0 && num <= 59) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][3*2]:indls[0][3*2+1]]), 10, 64)
		if !(num >= 0 && num <= 59) {
			return false
		}
		return true
	}
	return false
}

func IsTime(word []byte) bool {
	numberre := regexp.MustCompile("(?ism)^([0-9]{2}):([0-9]{2})$")
	indls := numberre.FindAllSubmatchIndex([]byte(word), -1)
	if len(indls) == 1 {
		num, _ := strconv.ParseInt(string(word[indls[0][1*2]:indls[0][1*2+1]]), 10, 64)
		if !(num >= 0 && num <= 23) {
			return false
		}
		num, _ = strconv.ParseInt(string(word[indls[0][2*2]:indls[0][2*2+1]]), 10, 64)
		if !(num >= 0 && num <= 59) {
			return false
		}
		return true
	}
	return false
}

//for + - * / * () sin cos tan asin acos atan sum avg 。。。 ^ {} []
//type float64,array,namemap,string,wav,jpg
func DoPartCalc(expr []byte, expri *int, ends []string, stack *sync.Map, exprcheck bool) (rlval interface{}, ok bool) {
	if *expri < len(expr) {
		switch expr[*expri] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+': //until + - ) } ]
			numberre := regexp.MustCompile("(?ism)^([+-]?[0-9]+(\\.[0-9]+)?)([Ee][+-]?[0-9]+(\\.[0-9]+)?)?")
			inds := numberre.FindAllSubmatchIndex(expr, -1)
			if len(inds) > 0 {
				num1 := string(expr[*expri+inds[0][0] : *expri+inds[0][1]])
				*expri += inds[0][1]
				for len(ends) == 0 || *expri < len(expr) && SliceLastIndex(ends, string([]byte{expr[*expri]})) == -1 { //until [end] or end in ends
					//operate could:+ - * / , ) ]
					var op1 byte
					if *expri < len(expr) {
						op1 = expr[*expri]
						*expri += 1
					} else {
						val, ok := stack.Load("")
						if ok {
							return val, true
						}
						return nil, false
					}
					if op1 == '+' || op1 == '-' || op1 == '*' || op1 == '/' {
						// func num ( [ { [end]
						if *expri >= len(expr) {
							panic("error end")
						}
						var num2 string
						var op2 byte
						inds2 := numberre.FindAllSubmatchIndex(expr[*expri:], -1)
						if len(inds2) > 0 {
							num2 = string(expr[*expri+inds2[0][0] : *expri+inds2[0][1]])
							*expri += inds2[0][1]
							//operate could:+ - * / , ) ]
							if *expri < len(expr) {
								op2 = expr[*expri]
							}
						} else {
							//maybe [end] ( [ { function
							if unicode.IsLetter(rune(expr[*expri])) {
								rl, ok := DoPartCalc(expr, expri, []string{")"}, stack, exprcheck)
								if ok != true {
									return nil, false
								}
								num2 = rl.(string)
								if *expri < len(expr) {
									op2 = expr[*expri]
								}
							} else if expr[*expri] == '(' {
								rl, ok := DoPartCalc(expr, expri, []string{")"}, stack, exprcheck)
								if ok != true {
									return nil, false
								}
								num2 = rl.(string)
								if *expri < len(expr) {
									op2 = expr[*expri]
								}
							} else if expr[*expri] == '{' {
								rl, ok := DoPartCalc(expr, expri, []string{"}"}, stack, exprcheck)
								if ok != true {
									return nil, false
								}
								num2 = rl.(string)
								if *expri < len(expr) {
									op2 = expr[*expri]
								}
							} else if expr[*expri] == '[' {
								rl, ok := DoPartCalc(expr, expri, []string{"]"}, stack, exprcheck)
								if ok != true {
									return nil, false
								}
								num2 = rl.(string)
								if *expri < len(expr) {
									op2 = expr[*expri]
								}
							}
						}
						if op1 == '*' || op1 == '/' {
							if exprcheck == false {
								crl, crle := method.Call(string([]byte{op1}), string(num1), string(num2))
								if crle != nil {
									return nil, false
								}
								num1 = crl[0].String()
							} else {
								num1 = "1"
							}
							(*stack).Store("", num1)
							op1 = op2
							continue
						} else {
							var num3 string
							if op2 == '*' || op2 == '/' {
								*expri += inds2[0][1]
								retval, ok2 := DoPartCalc(expr, expri, ends, stack, exprcheck)
								if !ok2 {
									return nil, false
								}
								num3 = retval.(string)
								if exprcheck == false {
									frl, frle := method.Call(string([]byte{op2}), string(num2), string(num3))
									if frle != nil {
										return nil, false
									}
									num4 := frl[0].String()
									frl, frle = method.Call(string([]byte{op1}), string(num1), string(num4))
									if frle != nil {
										return nil, false
									}
									num1 = frl[0].String()
								} else {
									num1 = "1"
								}
								(*stack).Store("", num1)
								continue
							} else {
								if op1 == '+' {
									if exprcheck == false {
										crl, crle := method.Call("jia", string(num1), string(num2))
										if crle != nil {
											return nil, false
										}
										num1 = crl[0].String()
									} else {
										num1 = "1"
									}
									(*stack).Store("", num1)
									continue
								} else if op1 == '-' {
									if exprcheck == false {
										crl, crle := method.Call("jianr", string(num1), string(num2))
										if crle != nil {
											return nil, false
										}
										num1 = crl[0].String()
									} else {
										num1 = "1"
									}
									(*stack).Store("", num1)
									continue
								}
							}
						}
					} else {
						//maybe:),},],,,[end]
						//return
						val, ok := stack.Load("")
						if ok {
							return val, true
						}
						return nil, false
					}
				}
				val, ok := stack.Load("")
				if ok {
					return val, true
				}
				return nil, false
			}
		case '[':
			*expri += 1
			val, vale := DoPartCalc(expr, expri, []string{"]"}, stack, exprcheck)
			if vale != true {
				return nil, false
			}
			if *expri < len(expr) && expr[*expri] != ')' {
				return nil, false
			}
			*expri += 1 //skip ]
			return val, true
		case '{':
			*expri += 1
			val, vale := DoPartCalc(expr, expri, []string{"}"}, stack, exprcheck)
			if vale != true {
				return nil, false
			}
			if *expri < len(expr) && expr[*expri] != ')' {
				return nil, false
			}
			*expri += 1 //skip }
			return val, true
		case 'a', 'b', 'c', 'd', 'e', 'f', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
			//until + - ) ] }
			funchead := ""
			if *expri >= len(expr) {
				val, ok := stack.Load("")
				if ok {
					return val, true
				}
				return nil, false
			}
			//fmt.Println("pppppppppp", *expri, len(expr), string(expr))
			for *expri < len(expr) && (unicode.IsLetter(rune(expr[*expri])) || unicode.IsDigit(rune(expr[*expri]))) {
				funchead += string(expr[*expri : *expri+1])
				*expri += 1
			}
			if *expri < len(expr) && expr[*expri] != '(' || *expri >= len(expr) {
				return nil, false
			}
			//funchead += "("
			*expri += 1 //skip func (
			//prepare paramer list
			*expri += 1
			val, vale := DoPartCalc(expr, expri, []string{")"}, stack, exprcheck)
			if vale != true {
				return nil, false
			}
			if *expri < len(expr) && expr[*expri] != ')' {
				return nil, false
			}
			*expri += 1 //skip )
			//invoke
			if exprcheck == false {
				frl, frle := method.Call(funchead, val)
				if frle != nil {
					return nil, false
				}
				return frl[0].String(), true
			} else {
				return "1", true
			}
			//return
		case '(':
			//until current end )
			*expri += 1
			crl, crle := DoPartCalc(expr, expri, []string{")"}, stack, exprcheck)
			if crle != true {
				return nil, false
			}
			if *expri >= len(expr) || expr[*expri] != ')' {
				return nil, false
			}
			*expri += 1 //skip )
			return crl, true
		}
	}
	return nil, false
}

func CombinationDo(f func([]string) bool, params ...[]string) {
	paramsisize := []int{}
	paramsicur := []int{}
	for i := 0; i < len(params); i++ {
		paramsisize = append(paramsisize, len(params[i]))
		paramsicur = append(paramsicur, 0)
	}
	paramls := []string{}
	for i := 0; i < len(params); i++ {
		paramsicur[i] = 0
		paramls = append(paramls, params[i][0])
	}
	paramiend := len(params) - 1
	paramsicur[paramiend] = -1
	for true {
		bchangeiend := false
		for paramsicur[paramiend]+1 >= paramsisize[paramiend] {
			paramiend -= 1
			if paramiend < 0 {
				return
			}
			bchangeiend = true
		}
		if bchangeiend {
			paramsicur[paramiend] += 1
			paramls = paramls[:paramiend]
			for i := paramiend; i < len(params); i++ {
				if i > paramiend {
					paramsicur[i] = 0
				}
				paramls = append(paramls, params[i][paramsicur[i]])
			}
			paramiend = len(params) - 1
		} else {
			paramsicur[paramiend] += 1
			paramls[paramiend] = params[paramiend][paramsicur[paramiend]]
		}
		rl := f(paramls)
		if rl == false {
			return
		}
	}
}

func PermutationDo(f func(map[string]int) bool, params ...[]string) {
	paramsisize := []int{}
	paramsicur := []int{}
	for i := 0; i < len(params); i++ {
		paramsisize = append(paramsisize, len(params[i]))
		paramsicur = append(paramsicur, 0)
	}
	paramls := make(map[string]int, 0)
	for i := 0; i < len(params); i++ {
		paramsicur[i] = 0

		_, bold := paramls[params[i][paramsicur[i]]]
		if bold {
			paramls[params[i][paramsicur[i]]] += 1
		} else {
			paramls[params[i][paramsicur[i]]] = 1
		}
		if i+1 == len(params) {
			paramls[params[i][paramsicur[i]]] -= 1
		}
	}
	paramiend := len(params) - 1
	paramsicur[paramiend] = -1
	for true {
		bchangeiend := false
		for paramsicur[paramiend]+1 >= paramsisize[paramiend] {
			if paramls[params[paramiend][paramsicur[paramiend]]]-1 <= 0 {
				delete(paramls, params[paramiend][paramsicur[paramiend]])
			} else {
				paramls[params[paramiend][paramsicur[paramiend]]] -= 1
			}
			paramiend -= 1
			if paramiend < 0 {
				return
			}
			bchangeiend = true
		}
		if bchangeiend {
			if paramsicur[paramiend] >= 0 {
				//fmt.Println("paramiend", paramiend)
				if paramls[params[paramiend][paramsicur[paramiend]]]-1 <= 0 {
					delete(paramls, params[paramiend][paramsicur[paramiend]])
					//fmt.Println("paramls1", paramls)
				} else {
					paramls[params[paramiend][paramsicur[paramiend]]] -= 1
					//fmt.Println("paramls2", paramls)
				}
			}
			paramsicur[paramiend] += 1
			for i := paramiend; i < len(params); i++ {
				if i > paramiend {
					paramsicur[i] = 0
				}
				//fmt.Println("first", paramiend, params[i][paramsicur[i]])
				_, bold := paramls[params[i][paramsicur[i]]]
				if bold {
					paramls[params[i][paramsicur[i]]] += 1
				} else {
					paramls[params[i][paramsicur[i]]] = 1
				}
			}
			paramiend = len(params) - 1
		} else {
			if paramsicur[paramiend] >= 0 {
				if paramls[params[paramiend][paramsicur[paramiend]]]-1 <= 0 {
					delete(paramls, params[paramiend][paramsicur[paramiend]])
				} else {
					paramls[params[paramiend][paramsicur[paramiend]]] -= 1
				}
			}
			paramsicur[paramiend] += 1
			_, bold := paramls[params[paramiend][paramsicur[paramiend]]]
			if bold {
				paramls[params[paramiend][paramsicur[paramiend]]] += 1
			} else {
				paramls[params[paramiend][paramsicur[paramiend]]] = 1
			}
		}
		//fmt.Println(paramls)
		if len(paramls) == len(params) {
			rl := f(paramls)
			if rl == false {
				return
			}
		}
	}
}

func StringListHaveRepeat(list []string) bool {
	lmap := make(map[string]int8, 0)
	for _, val2 := range list {
		lmap[val2] = 1
	}
	if len(lmap) != len(list) {
		return true
	}
	return false
}

func StringListIndex(strls []string, val string) (index int) {
	for i := 0; i < len(strls); i++ {
		if strls[i] == val {
			return i
		}
	}
	return -1
}

func StringListCheckRepeatWithout(list, withoutval []string, withoutshowsametime, withoutgreatthan1, comeupsametime [][]string) bool {
	lmap := make(map[string]int, 0)
	emptycnt := 0
	for _, val2 := range list {
		if StringListIndex(withoutval, val2) != -1 {
			emptycnt += 1
			continue
		}
		_, bok := lmap[val2]
		if bok == false {
			lmap[val2] = 1
		} else {
			lmap[val2] += 1
		}
	}
	if len(lmap)+emptycnt != len(list) {
		return true
	}
	for i := 0; i < len(withoutshowsametime); i++ {
		okcnt := 0
		for j := 0; j < len(withoutshowsametime[i]); j++ {
			_, bok := lmap[withoutshowsametime[i][j]]
			if bok == false {
				break
			} else {
				okcnt += 1
			}
		}
		if okcnt == len(withoutshowsametime[i]) {
			return true
		}
	}
	for i := 0; i < len(withoutgreatthan1); i++ {
		okcnt := 0
		for j := 0; j < len(withoutgreatthan1[i]); j++ {
			cnt, bok := lmap[withoutgreatthan1[i][j]]
			if bok {
				okcnt += cnt
			}
		}
		if okcnt > 1 {
			return true
		}
	}
	for i := 0; i < len(comeupsametime); i++ {
		okcnt := 0
		for j := 0; j < len(comeupsametime[i]); j++ {
			_, bok := lmap[comeupsametime[i][j]]
			if bok == false {
				break
			} else {
				okcnt += 1
			}
		}
		if okcnt > 0 && okcnt != len(withoutshowsametime[i]) {
			return true
		}
	}
	return false
}

func StringListHaveRepeatExceptEmptyOneSpace(list []string) bool {
	lmap := make(map[string]int8, 0)
	emptycnt := 0
	for _, val2 := range list {
		if val2 == "" || val2 == " " {
			emptycnt += 1
			continue
		}
		lmap[val2] = 1
	}
	if len(lmap)+emptycnt != len(list) {
		return true
	}
	return false
}

func HaveRepeat(list interface{}) bool {
	switch list.(type) {
	case []int:
		{
			lmap := make(map[int]int8, 0)
			for _, val2 := range list.([]int) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]int)) {
				return true
			}
		}
	case []int8:
		{
			lmap := make(map[int8]int8, 0)
			for _, val2 := range list.([]int8) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]int8)) {
				return true
			}
		}
	case []int16:
		{
			lmap := make(map[int16]int8, 0)
			for _, val2 := range list.([]int16) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]int16)) {
				return true
			}
		}
	case []int32:
		{
			lmap := make(map[int32]int8, 0)
			for _, val2 := range list.([]int32) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]int32)) {
				return true
			}
		}
	case []int64:
		{
			lmap := make(map[int64]int8, 0)
			for _, val2 := range list.([]int64) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]int64)) {
				return true
			}
		}
	case []uint8:
		{
			lmap := make(map[uint8]int8, 0)
			for _, val2 := range list.([]uint8) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]uint8)) {
				return true
			}
		}
	case []uint16:
		{
			lmap := make(map[uint16]int8, 0)
			for _, val2 := range list.([]uint16) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]uint16)) {
				return true
			}
		}
	case []uint32:
		{
			lmap := make(map[uint32]int8, 0)
			for _, val2 := range list.([]uint32) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]uint32)) {
				return true
			}
		}
	case []uint64:
		{
			lmap := make(map[uint64]int8, 0)
			for _, val2 := range list.([]uint64) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]uint64)) {
				return true
			}
		}
	case []string:
		{
			lmap := make(map[string]int8, 0)
			for _, val2 := range list.([]string) {
				lmap[val2] = 1
			}
			if len(lmap) != len(list.([]string)) {
				return true
			}
		}
	case [][]byte:
		{
			lmap := make(map[string]int8, 0)
			for _, val2 := range list.([][]byte) {
				lmap[string(val2)] = 1
			}
			if len(lmap) != len(list.([][]byte)) {
				return true
			}
		}
	}
	return false
}

func GetTagContent(text, tagstart, tagend string) (tagcttls []string) {
	if tagstart == tagend || tagstart == "" || tagend == "" {
		return nil
	}
	i := 0
	for {
		j := strings.Index(text[i:], tagstart)
		if j != -1 {
			i2 := strings.Index(text[i:], tagend)
			if i2 != -1 {
				tagcttls = append(tagcttls, text[i+j+len(tagstart):i+i2])
				i = i + i2 + len(tagend)
			} else {
				break
			}
		} else {
			break
		}
	}
	return tagcttls
}

func GetRecursiveTagContent(text, tagmark, tagstart, tagend string) (tagcttls []string) {
	if strings.Compare(tagmark[:len(tagstart)], tagstart) != 0 {
		return nil
	}
	i := strings.Index(text, tagmark)
	if i == -1 || tagstart == tagend || tagstart == "" || tagend == "" {
		return nil
	}
	deepstart := 0
	deep := 0
	for {
		j := strings.Index(text[i:], tagstart)
		i2 := strings.Index(text[i:], tagend)
		if i2 != -1 && i2 < j {
			if deep-1 >= 0 {
				deep -= 1
				i = i + i2 + len(tagend)
				continue
			}
		}
		if j != -1 {
			i3 := strings.Index(text[i:], tagstart)
			if i2 != -1 {
				if (i3 > i2 || i3 == -1) && deep == 0 {
					fmt.Println("s e", deepstart, i+i2)
					tagcttls = append(tagcttls, text[deepstart:i+i2])
					i = i + i2 + len(tagend)
				} else {
					if deep == 0 {
						deepstart = i + j + len(tagstart)
					}
					deep += 1
					i = i + i3 + len(tagstart)
				}
			} else {
				break
			}
		} else {
			if i2 != -1 {
				deep -= 1
				if deep == 0 {
					fmt.Println("s e", deepstart, i+i2)
					tagcttls = append(tagcttls, text[deepstart:i+i2])
					i = i + i2 + len(tagend)
				}
				i = i + i2 + len(tagend)
				continue
			}
			break
		}
	}
	return tagcttls
}

func GetAndTruncateTagContent(text, tagstart, tagend string) (outtext string, tagcttls []string) {
	if tagstart == tagend || tagstart == "" || tagend == "" {
		return text, nil
	}
	i := 0
	posls := []int{}
	for {
		j := strings.Index(text[i:], tagstart)
		if j != -1 {
			i2 := strings.Index(text[i:], tagend)
			if i2 != -1 {
				posls = append(posls, i+j+len(tagstart))
				posls = append(posls, i+i2)
				tagcttls = append(tagcttls, text[i+j+len(tagstart):i+i2])
				i = i + i2 + len(tagend)
			} else {
				break
			}
		} else {
			break
		}
	}
	if len(posls) > 0 {
		outtext += text[:posls[0]-len(tagstart)]
		for i := 2; i < len(posls); i += 2 {
			fmt.Println(posls[i-1]+len(tagend), posls[i]-len(tagstart))
			outtext += text[posls[i-1]+len(tagend) : posls[i]-len(tagstart)]
		}
		outtext += text[posls[len(posls)-1]+len(tagend):]
		return outtext, tagcttls
	} else {
		return text, nil
	}
}

func FindTagContentPosition(text, tagstart, tagend string) (poslist []int) {
	if tagstart == tagend || tagstart == "" || tagend == "" {
		return nil
	}
	i := 0
	posls := []int{}
	for {
		j := strings.Index(text[i:], tagstart)
		if j != -1 {
			i2 := strings.Index(text[i:], tagend)
			if i2 != -1 {
				posls = append(posls, i+j+len(tagstart))
				posls = append(posls, i+i2)
				i = i + i2 + len(tagend)
			} else {
				break
			}
		} else {
			break
		}
	}
	return posls
}

func ListUniqueAdd(ls []string, value string) (ls2 []string) {
	ind := SliceSearch(ls, value, 0)
	if ind == -1 {
		ls2 = append(ls, value)
		return ls2
	}
	return ls
}

func ListMapGet(allgroupid, allgrouplastview []string, groupid string) string {
	ind := SliceSearch(allgroupid, groupid, 0)
	if ind == -1 {
		return ""
	}
	return allgrouplastview[ind]
}

func ListMapSet(allgroupid, allgrouplastview []string, groupid, value string) bool {
	ind := SliceSearch(allgroupid, groupid, 0)
	if ind == -1 {
		return false
	}
	allgrouplastview[ind] = value
	return true
}

func IsoTimeToTime(ti string) time.Time {
	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05", ti, time.Local)
	return todayZero
}

func IsoDateSecondTimeCompare(lastrefreshtime, msgtime string) int {
	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05", lastrefreshtime, time.Local)
	todayZero2, _ := time.ParseInLocation("2006-01-02 15:04:05", msgtime, time.Local)
	if todayZero.Before(todayZero2) {
		return -1
	} else if todayZero.Equal(todayZero2) {
		return 0
	} else {
		return 1
	}
}

func IsoDateMilisecTimeCompare(lastrefreshtime, msgtime string) int {
	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05.000", lastrefreshtime, time.Local)
	todayZero2, _ := time.ParseInLocation("2006-01-02 15:04:05.000", msgtime, time.Local)
	if todayZero.Before(todayZero2) {
		return -1
	} else if todayZero.Equal(todayZero2) {
		return 0
	} else {
		return 1
	}
}

func IsoDateMicrosecTimeCompare(lastrefreshtime, msgtime string) int {
	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05.000000", lastrefreshtime, time.Local)
	todayZero2, _ := time.ParseInLocation("2006-01-02 15:04:05.000000", msgtime, time.Local)
	if todayZero.Before(todayZero2) {
		return -1
	} else if todayZero.Equal(todayZero2) {
		return 0
	} else {
		return 1
	}
}

func IsoDateNanosecTimeCompare(lastrefreshtime, msgtime string) int {
	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05.000000000", lastrefreshtime, time.Local)
	todayZero2, _ := time.ParseInLocation("2006-01-02 15:04:05.000000000", msgtime, time.Local)
	if todayZero.Before(todayZero2) {
		return -1
	} else if todayZero.Equal(todayZero2) {
		return 0
	} else {
		return 1
	}
}

// func TimeCompare3mm(lastrefreshtime, msgtime string) int {
// 	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05.000", lastrefreshtime, time.Local)
// 	todayZero2, _ := time.ParseInLocation("2006-01-02 15:04:05.000", msgtime, time.Local)
// 	if todayZero.Before(todayZero2) {
// 		return -1
// 	} else if todayZero.Equal(todayZero2) {
// 		return 0
// 	} else {
// 		return 1
// 	}
// }

// func TimeCompare6mm(lastrefreshtime, msgtime string) int {
// 	todayZero, _ := time.ParseInLocation("2006-01-02 15:04:05.000000", lastrefreshtime, time.Local)
// 	todayZero2, _ := time.ParseInLocation("2006-01-02 15:04:05.000000", msgtime, time.Local)
// 	if todayZero.Before(todayZero2) {
// 		return -1
// 	} else if todayZero.Equal(todayZero2) {
// 		return 0
// 	} else {
// 		return 1
// 	}
// }

func StringListToBytes(strls []string) (nodebytes []byte) {
	nodebytes = make([]byte, 0, 2048)
	for i := 0; i < len(strls); i++ {
		nodebytes = append(nodebytes, PackUint32(uint32(len(strls[i])))...)
		nodebytes = append(nodebytes, []byte(strls[i])...)
	}
	return nodebytes
}

func StringListFromBytes(nodebytes []byte) (strls []string) {
	strls = make([]string, 0, 256)
	nodepos := uint32(0)
	for nodepos < uint32(len(nodebytes)) {
		itemlen := UnpackUint32(nodebytes[nodepos : nodepos+4])
		nodepos += 4
		strls = append(strls, string(nodebytes[nodepos:nodepos+itemlen]))
		nodepos += itemlen
	}
	return strls
}

func PngResize(pathfrom, pathto string, width uint) {
	file, err := os.Open(pathfrom)
	if err != nil {
		log.Fatal(err)
	}

	// decode jpeg into image.Image
	img, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(width, 0, img, resize.NearestNeighbor)

	out, err := os.Create(pathto)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	png.Encode(out, m)
}

func JpegResize(pathfrom, pathto string, width uint) {
	file, err := os.Open(pathfrom)
	if err != nil {
		log.Fatal(err)
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	// resize to width 1000 using Lanczos resampling
	// and preserve aspect ratio
	m := resize.Resize(width, 0, img, resize.Lanczos3)

	out, err := os.Create(pathto)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}

//support jpg png
func ImageResize(pathfrom, pathto string, width uint) {
	var m image.Image
	if strings.HasSuffix(pathfrom, ".png") {
		file, err := os.Open(pathfrom)
		if err != nil {
			log.Fatal(err)
		}

		// decode jpeg into image.Image
		img, err := png.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		// resize to width 1000 using Lanczos resampling
		// and preserve aspect ratio
		m = resize.Resize(width, 0, img, resize.NearestNeighbor)
	} else if strings.HasSuffix(pathfrom, ".jpg") {
		file, err := os.Open(pathfrom)
		if err != nil {
			log.Fatal(err)
		}

		// decode jpeg into image.Image
		img, err := jpeg.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()

		// resize to width 1000 using Lanczos resampling
		// and preserve aspect ratio
		m = resize.Resize(width, 0, img, resize.Lanczos3)
	}

	out, err := os.Create(pathto)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
}

//support jpg png
func ImageResizeBig(pathfrom, pathto string, width uint) (result bool) {
	defer func() {
		for r := recover(); r != nil; {
			fmt.Println("r", r)
			fmt.Println("fomt error")
			result = false
			r = recover()
		}
	}()
	ff, ffe := os.OpenFile(pathfrom, os.O_RDONLY, 0666)
	if ffe != nil {
		return false
	} else {
		bt1024 := make([]byte, 1024)
		ff.Read(bt1024)
		ff.Close()
		mime := MimeFromIncipit(bt1024)
		if mime == "image/jpeg" {
			if !strings.HasSuffix(pathfrom, ".jpg") {
				if strings.LastIndex(pathfrom, ".") != -1 {
					os.Rename(pathfrom, pathfrom[:strings.LastIndex(pathfrom, ".")]+".jpg")
					pathfrom = pathfrom[:strings.LastIndex(pathfrom, ".")] + ".jpg"
				} else {
					os.Rename(pathfrom, pathfrom+".jpg")
					pathfrom += ".jpg"
				}
			}
		} else if mime == "image/png" {
			if !strings.HasSuffix(pathfrom, ".png") {
				if strings.LastIndex(pathfrom, ".") != -1 {
					os.Rename(pathfrom, pathfrom[:strings.LastIndex(pathfrom, ".")]+".png")
					pathfrom = pathfrom[:strings.LastIndex(pathfrom, ".")] + ".png"
				} else {
					os.Rename(pathfrom, pathfrom+".png")
					pathfrom += ".png"
				}
			}
		} else {
			return false
		}
	}

	var m image.Image
	if strings.HasSuffix(pathfrom, ".png") {
		file, err := os.Open(pathfrom)
		if err != nil {
			return false
		}

		// decode jpeg into image.Image
		img, err := png.Decode(file)
		if err != nil {
			return false
		}
		file.Close()

		// resize to width 1000 using Lanczos resampling
		// and preserve aspect ratio

		if img.Bounds().Dx() > int(width) {
			m = resize.Resize(width, 0, img, resize.NearestNeighbor)
		} else {
			m = img
		}
	} else if strings.HasSuffix(pathfrom, ".jpg") {
		file, err := os.Open(pathfrom)
		if err != nil {
			return false
		}

		// decode jpeg into image.Image
		img, err := jpeg.Decode(file)
		if err != nil {
			return false
		}
		file.Close()

		// resize to width 1000 using Lanczos resampling
		// and preserve aspect ratio
		if img.Bounds().Dx() > int(width) {
			m = resize.Resize(width, 0, img, resize.Lanczos3)
		} else {
			m = img
		}
	}

	out, err := os.Create(pathto)
	if err != nil {
		return false
	}
	defer out.Close()

	// write new image to file
	jpeg.Encode(out, m, nil)
	return true
}

// image formats and magic numbers
var magicTable = map[string]string{
	"\xff\xd8\xff":      "image/jpeg",
	"\x89PNG\r\n\x1a\n": "image/png",
	"GIF87a":            "image/gif",
	"GIF89a":            "image/gif",
}

// mimeFromIncipit returns the mime type of an image file from its first few
// bytes or the empty string if the file does not look like a known file type
func MimeFromIncipit(incipit []byte) string {
	incipitStr := []byte(incipit)
	for magic, mime := range magicTable {
		if strings.HasPrefix(string(incipitStr), magic) {
			return mime
		}
	}

	return ""
}

func GetRepeatCountList(strls []string) (countls []int) {
	countls = make([]int, 0)
	if len(strls) == 0 {
		return countls
	}
	countls = append(countls, 1)
	curpos := 0
	strlsi := 0
	prestr := strls[0]
	strlsi += 1
	for strlsi < len(strls) {
		if prestr != strls[strlsi] {
			prestr = strls[strlsi]
			countls = append(countls, 1)
			curpos += 1
		} else {
			countls[curpos] += 1
		}
		strlsi += 1
	}
	return countls
}

//return unix nanao
func GetFileModifiedNanoTime(filepath string) int64 {
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("open file error")
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Println("stat fileinfo error")
		return time.Now().Unix()
	}

	return fi.ModTime().UnixNano()
}

func GetFileModifiedMicroTime(filepath string) int64 {
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("open file error")
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Println("stat fileinfo error")
		return time.Now().Unix()
	}

	return fi.ModTime().UnixNano() / 1000
}

func CopyStringSlice(strls []string) []string {
	mm := make([]string, len(strls))
	for i := 0; i < len(mm); i++ {
		mm[i] = strls[i]
	}
	return mm
}

func FileSha1(filepath string) (sh1str string) {
	ff, ffe := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if ffe == nil {
		npos, npose := ff.Seek(0, os.SEEK_END)
		if npose == nil {
			ff.Seek(0, os.SEEK_SET)
			buf := make([]byte, 102400)
			sh1 := sha1.New()
			for i := int64(0); i < npos; i += 102400 {
				n, ne := ff.Read(buf)
				if ne != nil {
					break
				}
				sh1.Write(buf[:n])
			}
			sh1str = fmt.Sprintf("%x", sh1.Sum(nil))
		}
		ff.Close()
	}
	return sh1str
}

type KeyValueBuf struct {
	buf    []byte
	curpos int
	key    string
	value  string
}

func NewKeyValueBuf(raw []byte) *KeyValueBuf {
	kvbuf := &KeyValueBuf{}
	if raw != nil {
		kvbuf.buf = raw
		kvbuf.curpos = len(raw)
	}
	return kvbuf
}

func (kvbuf *KeyValueBuf) Append(key, val string) bool {
	kvbuf.key = key
	kvbuf.value = val
	kvbuf.buf = append(kvbuf.buf, PackUint32(uint32(len([]byte(key))))...)
	kvbuf.buf = append(kvbuf.buf, []byte(key)...)
	kvbuf.buf = append(kvbuf.buf, PackUint32(uint32(len([]byte(val))))...)
	kvbuf.buf = append(kvbuf.buf, []byte(val)...)
	kvbuf.curpos = len(kvbuf.buf)
	return true
}

func (kvbuf *KeyValueBuf) Reset() bool {
	kvbuf.curpos = 0
	return true
}

func (kvbuf *KeyValueBuf) Next() bool {
	if kvbuf.curpos != len(kvbuf.buf) {
		keylen := UnpackUint32(kvbuf.buf[kvbuf.curpos : kvbuf.curpos+4])
		keystart := kvbuf.curpos + 4
		kvbuf.curpos += 4 + int(keylen)
		if kvbuf.curpos >= len(kvbuf.buf) {
			return false
		}
		valuelen := UnpackUint32(kvbuf.buf[kvbuf.curpos : kvbuf.curpos+4])
		valstart := kvbuf.curpos + 4
		kvbuf.curpos += 4 + int(valuelen)
		kvbuf.key = string(kvbuf.buf[keystart : keystart+int(keylen)])
		kvbuf.value = string(kvbuf.buf[valstart : valstart+int(valuelen)])
		return true
	}
	return false
}

func (kvbuf *KeyValueBuf) Key() string {
	return kvbuf.key
}

func (kvbuf *KeyValueBuf) Value() string {
	return kvbuf.value
}

func (kvbuf *KeyValueBuf) Buffer() []byte {
	return kvbuf.buf
}

func StringListClone(strls []string) (cpstrls []string) {
	cpstrls = make([]string, len(strls))
	for i := 0; i < len(strls); i++ {
		cpstrls[i] = strls[i]
	}
	return cpstrls
}

func StringList2DClone(strls [][]string) (cpstrls [][]string) {
	cpstrls = make([][]string, len(strls))
	for i := 0; i < len(strls); i++ {
		cpstrls[i] = make([]string, len(strls[i]))
		for j := 0; j < len(strls[i]); j++ {
			cpstrls[i][j] = strls[i][j]
		}
	}
	return cpstrls
}

func StringSetAdd(strset []string, str string) []string {
	if SliceSearch(strset, str, 0) == -1 {
		return append(strset, str)
	} else {
		return strset
	}
}

func StringSetRemove(strset []string, str string) []string {
	ind := SliceSearch(strset, str, 0)
	if ind != -1 {
		return append(strset[:ind], strset[ind+1:]...)
	} else {
		return strset
	}
}

func DecListToBytes(declist string) []byte {
	ls := regexp.MustCompile("[ ,]+").Split(declist, -1)
	var bt []byte
	for i := 0; i < len(ls); i++ {
		if len(ls[i]) == 0 {
			continue
		}
		val, vale := strconv.ParseInt(ls[i], 10, 32)
		if vale == nil {
			bt = append(bt, byte(val))
		}
	}
	return bt
}

func DecListToString(declist string) string {
	ls := regexp.MustCompile("[ ,]+").Split(declist, -1)
	var bt []byte
	for i := 0; i < len(ls); i++ {
		if len(ls[i]) == 0 {
			continue
		}
		val, vale := strconv.ParseInt(ls[i], 10, 32)
		if vale == nil {
			bt = append(bt, byte(val))
		}
	}
	return string(bt)
}

func FirstUCharLen(data []byte, startpos int) int {
	var utf8_byte_mask uint32 = 0x3f
	datai := startpos
	var unichar uint32
	unicharoplen := 0
	datalen := len(data) - startpos
	for datai < len(data) {
		lead := uint32(data[datai])
		if lead < 0x80 {
			// 0xxxxxxx -> U+0000..U+007F
			unichar = lead
			if unichar < 0x80 {
				// U+0000..U+007F
				unicharoplen = 1
			} else if unichar < 0x800 {
				// U+0080..U+07FF
				unicharoplen = 2
			} else {
				// U+0800..U+FFFF
				unicharoplen = 3
			}
			return unicharoplen
		} else if uint32(lead-0xC0) < 0x20 && datalen >= 2 && (data[datai+1]&0xc0) == 0x80 {
			// 110xxxxx -> U+0080..U+07FF
			unichar = ((lead & ^uint32(0xC0)) << 6) | (uint32(data[datai+1]) & utf8_byte_mask)
			if unichar < 0x80 {
				// U+0000..U+007F
				unicharoplen = 1
			} else if unichar < 0x800 {
				// U+0080..U+07FF
				unicharoplen = 2
			} else {
				// U+0800..U+FFFF
				unicharoplen = 3
			}
			return unicharoplen
		} else if (lead-0xE0) < 0x10 && datalen >= 3 && (data[datai+1]&0xc0) == 0x80 && (data[datai+2]&0xc0) == 0x80 {
			// 1110xxxx -> U+0800-U+FFFF
			unichar = ((lead & ^uint32(0xE0)) << 12) | ((uint32(data[datai+1]) & utf8_byte_mask) << 6) | (uint32(data[datai+2]) & utf8_byte_mask)
			if unichar < 0x80 {
				// U+0000..U+007F
				unicharoplen = 1
			} else if unichar < 0x800 {
				// U+0080..U+07FF
				unicharoplen = 2
			} else {
				// U+0800..U+FFFF
				unicharoplen = 3
			}
			return unicharoplen
		} else if (lead-0xF0) < 0x08 && datalen >= 4 && (data[datai+1]&0xc0) == 0x80 && (data[datai+2]&0xc0) == 0x80 && (data[datai+3]&0xc0) == 0x80 {
			// 11110xxx -> U+10000..U+10FFFF
			unichar = ((lead & ^uint32(0xF0)) << 18) | ((uint32(data[datai+1]) & utf8_byte_mask) << 12) | ((uint32(data[datai+2]) & utf8_byte_mask) << 6) | (uint32(data[datai+3]) & utf8_byte_mask)
			unicharoplen = 4
			return unicharoplen
		} else {
			// 10xxxxxx or 11111xxx -> invalid
			datai++
			datalen -= 1
		}
	}

	return unicharoplen
}

func GetFileModTimeSecond(path string) int64 {
	f, err := os.Open(path)
	if err != nil {
		log.Println("open file error")
		return 0
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Println("stat fileinfo error")
		return 0
	}
	return fi.ModTime().Unix()
}

func IsoTimeToSecond(humanval string) int64 {
	var year, month, day, hour, minute, sec int64
	if strings.Index(humanval, "-") != -1 {
		year, _ = strconv.ParseInt(humanval[0:strings.Index(humanval, "-")], 10, 32)
		year -= 1970
		humanval = humanval[strings.Index(humanval, "-")+1:]
		if strings.Index(humanval, "-") != -1 {
			month, _ = strconv.ParseInt(humanval[0:strings.Index(humanval, "-")], 10, 32)
			humanval = humanval[strings.Index(humanval, "-")+1:]
			if strings.Index(humanval, " ") != -1 {
				day, _ = strconv.ParseInt(humanval[0:strings.Index(humanval, " ")], 10, 32)
				humanval = humanval[strings.Index(humanval, " ")+1:]
				if strings.Index(humanval, ":") != -1 {
					hour, _ = strconv.ParseInt(humanval[0:strings.Index(humanval, ":")], 10, 32)
					humanval = humanval[strings.Index(humanval, ":")+1:]
					if strings.Index(humanval, ":") != -1 {
						minute, _ = strconv.ParseInt(humanval[0:strings.Index(humanval, ":")], 10, 32)
						humanval = humanval[strings.Index(humanval, ":")+1:]
						sec, _ = strconv.ParseInt(humanval, 10, 32)
					} else {
						minute, _ = strconv.ParseInt(humanval, 10, 32)
					}
				}
			} else {
				day, _ = strconv.ParseInt(humanval, 10, 32)
			}
		}
	}
	isleapyear := false
	if year%4 == 0 {
		if year%100 == 0 {
			if year%400 == 0 {
				isleapyear = true
			}
		} else {
			isleapyear = true
		}
	}
	leapyear := (year - 1) / 4
	commonyear := (year - 1) / 400
	totalday := (leapyear-commonyear)*366 + (year-(leapyear-commonyear))*365
	totalmonthday := 0
	for mi := 1; mi < int(month); mi++ {
		monthday := 30
		if mi == 1 || mi == 3 || mi == 5 || mi == 7 || mi == 8 || mi == 10 || mi == 12 {
			monthday = 31
		}
		if mi == 2 {
			if isleapyear == true {
				monthday = 29
			} else {
				monthday = 28
			}
		}
		totalmonthday += monthday
	}

	totalday += int64(totalmonthday) + day
	totalSec := int64(totalday)*24*3600 + hour*3600 + minute*60 + sec
	return totalSec
}

func ValCompare(val1, val2 interface{}) int {
	if reflect.TypeOf(val1) == reflect.TypeOf(val2) {
		switch val1.(type) {
		case []string:
			{
				if String1DCompare(val1.([]string), val2.([]string)) {
					return 0
				} else {
					return -1
				}
			}
		case []interface{}:
			{
				if len(val1.([]interface{})) == len(val2.([]interface{})) {
					for i := 0; i < len(val1.([]interface{})); i++ {
						//fmt.Println(reflect.TypeOf(val1.([]interface{})[i]), reflect.TypeOf(val2.([]interface{})[i]))
						//fmt.Println(val1, val2)
						if reflect.TypeOf(val1.([]interface{})[i]) == reflect.TypeOf(val2.([]interface{})[i]) {

							switch val1.([]interface{})[i].(type) {
							case string:
								{
									if val1.([]interface{})[i].(string) != val2.([]interface{})[i].(string) {
										return -1
									}
								}
							case float32:
								{
									if val1.([]interface{})[i].(float32) != val2.([]interface{})[i].(float32) {
										return -1
									}
								}
							case float64:
								{
									if val1.([]interface{})[i].(float64) != val2.([]interface{})[i].(float64) {
										return -1
									}
								}
							case int:
								{
									//fmt.Println(val1, val2)
									if val1.([]interface{})[i].(int) != val2.([]interface{})[i].(int) {
										return -1
									}
								}
							case uint:
								{
									if val1.([]interface{})[i].(uint) != val2.([]interface{})[i].(uint) {
										return -1
									}
								}
							case int8:
								{
									if val1.([]interface{})[i].(int8) != val2.([]interface{})[i].(int8) {
										return -1
									}
								}
							case int16:
								{
									if val1.([]interface{})[i].(int16) != val2.([]interface{})[i].(int16) {
										return -1
									}
								}
							case int32:
								{
									if val1.([]interface{})[i].(int32) != val2.([]interface{})[i].(int32) {
										return -1
									}
								}
							case int64:
								{
									if val1.([]interface{})[i].(int64) != val2.([]interface{})[i].(int64) {
										return -1
									}
								}
							case uint8:
								{
									if val1.([]interface{})[i].(uint8) != val2.([]interface{})[i].(uint8) {
										return -1
									}
								}
							case uint16:
								{
									if val1.([]interface{})[i].(uint16) != val2.([]interface{})[i].(uint16) {
										return -1
									}
								}
							case uint32:
								{
									if val1.([]interface{})[i].(uint32) != val2.([]interface{})[i].(uint32) {
										return -1
									}
								}
							case uint64:
								{
									if val1.([]interface{})[i].(uint64) != val2.([]interface{})[i].(uint64) {
										return -1
									}
								}
							case bool:
								{
									if val1.([]interface{})[i].(bool) != val2.([]interface{})[i].(bool) {
										return -1
									}
								}
							default:
								{
									return -3
								}
							}
						} else {
							return -1
						}
					}
					return 0
				} else {
					return -1
				}

			}
		case string:
			{
				if val1.(string) != val2.(string) {
					return -1
				} else {
					return 0
				}
			}
		case float32:
			{
				if val1.(float32) != val2.(float32) {
					return -1
				} else {
					return 0
				}
			}
		case float64:
			{
				if val1.(float64) != val2.(float64) {
					return -1
				} else {
					return 0
				}
			}
		case int:
			{
				if val1.(int) != val2.(int) {
					return -1
				}
			}
		case uint:
			{
				if val1.(uint) != val2.(uint) {
					return -1
				}
			}
		case int8:
			{
				if val1.(int8) != val2.(int8) {
					return -1
				}
			}
		case int16:
			{
				if val1.(int16) != val2.(int16) {
					return -1
				}
			}
		case int32:
			{
				if val1.(int32) != val2.(int32) {
					return -1
				}
			}
		case int64:
			{
				if val1.(int64) != val2.(int64) {
					return -1
				}
			}
		case uint8:
			{
				if val1.(uint8) != val2.(uint8) {
					return -1
				}
			}
		case uint16:
			{
				if val1.(uint16) != val2.(uint16) {
					return -1
				}
			}
		case uint32:
			{
				if val1.(uint32) != val2.(uint32) {
					return -1
				}
			}
		case uint64:
			{
				if val1.(uint64) != val2.(uint64) {
					return -1
				}
			}
		case bool:
			{
				if val1.(bool) != val2.(bool) {
					return -1
				} else {
					return 0
				}
			}

		case []float32:
			{
				if len(val1.([]float32)) == len(val2.([]float32)) {
					for i := 0; i < len(val1.([]float32)); i++ {
						if val1.([]float32)[i] != val2.([]float32)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []float64:
			{
				if len(val1.([]float64)) == len(val2.([]float64)) {
					for i := 0; i < len(val1.([]float64)); i++ {
						if val1.([]float64)[i] != val2.([]float64)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []int:
			{
				if len(val1.([]int)) == len(val2.([]int)) {
					for i := 0; i < len(val1.([]int)); i++ {
						if val1.([]int)[i] != val2.([]int)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []uint:
			{
				if len(val1.([]uint)) == len(val2.([]uint)) {
					for i := 0; i < len(val1.([]uint)); i++ {
						if val1.([]uint)[i] != val2.([]uint)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []int8:
			{
				if len(val1.([]int8)) == len(val2.([]int8)) {
					for i := 0; i < len(val1.([]int8)); i++ {
						if val1.([]int8)[i] != val2.([]int8)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []int16:
			{
				if len(val1.([]int16)) == len(val2.([]int16)) {
					for i := 0; i < len(val1.([]int16)); i++ {
						if val1.([]int16)[i] != val2.([]int16)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []int32:
			{
				if len(val1.([]int32)) == len(val2.([]int32)) {
					for i := 0; i < len(val1.([]int32)); i++ {
						if val1.([]int32)[i] != val2.([]int32)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []int64:
			{
				if len(val1.([]int64)) == len(val2.([]int64)) {
					for i := 0; i < len(val1.([]int64)); i++ {
						if val1.([]int64)[i] != val2.([]int64)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []uint8:
			{
				if len(val1.([]uint8)) == len(val2.([]uint8)) {
					for i := 0; i < len(val1.([]uint8)); i++ {
						if val1.([]uint8)[i] != val2.([]uint8)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []uint16:
			{
				if len(val1.([]uint16)) == len(val2.([]uint16)) {
					for i := 0; i < len(val1.([]uint16)); i++ {
						if val1.([]uint16)[i] != val2.([]uint16)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []uint32:
			{
				if len(val1.([]uint32)) == len(val2.([]uint32)) {
					for i := 0; i < len(val1.([]uint32)); i++ {
						if val1.([]uint32)[i] != val2.([]uint32)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []uint64:
			{
				if len(val1.([]uint64)) == len(val2.([]uint64)) {
					for i := 0; i < len(val1.([]uint64)); i++ {
						if val1.([]uint64)[i] != val2.([]uint64)[i] {
							return -1
						}
					}
					return 0
				}
			}
		case []bool:
			{
				if len(val1.([]bool)) == len(val2.([]bool)) {
					for i := 0; i < len(val1.([]bool)); i++ {
						if val1.([]bool)[i] != val2.([]bool)[i] {
							return -1
						}
					}
					return 0
				}
			}
		}
	}
	return -2
}

func WordSetMinus(WordSrc, WordMinus sync.Map) (newworddeq sync.Map) {
	WordSrc.Range(func(key, val interface{}) bool {
		_, bexists := WordMinus.Load(key)
		if !bexists {
			newworddeq.Store(key, 1)
		}
		return true
	})
	return WordSrc
}

func BKDRHash(str []byte) uint64 {
	seed := uint64(131313)
	hash := uint64(0)
	for i := 0; i < len(str); i++ {
		hash = (hash * seed) + uint64(str[i])
	}
	return hash
}

func SDBMHash(str []byte) uint64 {
	hash := uint64(0)
	for i := 0; i < len(str); i++ {
		hash = uint64(str[i]) + (hash << 6) + (hash << 16) - hash
	}
	return hash
}

func FileToMapU64Bytes(path string) (mdata map[uint64][]byte) {
	mdata = make(map[uint64][]byte, 0)
	ff, ffe := os.OpenFile(path, os.O_RDONLY, 0666)
	if ffe == nil {
		tempbt := make([]byte, 8)
		var key, vallen uint64
		for true {
			rdn, rdne := ff.Read(tempbt[:8])
			if rdne != nil || rdn != 8 {
				break
			}
			key = binary.BigEndian.Uint64(tempbt[:8])
			rdn, rdne = ff.Read(tempbt[:4])
			if rdne != nil || rdn != 4 {
				break
			}
			vallen = uint64(binary.BigEndian.Uint32(tempbt[:4]))
			valuebt := make([]byte, vallen)
			rdn, rdne = ff.Read(valuebt)
			if rdne != nil || rdn != int(vallen) {
				break
			}
			mdata[key] = valuebt
		}
		ff.Close()
	}
	return mdata
}

func MapU64BytesToFile(mdata map[uint64][]byte, path string) {
	ff, ffe := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if ffe == nil {
		keybt := make([]byte, 8)
		vallenbt := make([]byte, 4)
		outbt := make([]byte, 0, 4*1024*1024)
		var vallen uint32
		for key, value := range mdata {
			binary.BigEndian.PutUint64(keybt, key)
			vallen = uint32(len(value))
			binary.BigEndian.PutUint32(vallenbt, vallen)
			outbt = append(outbt, keybt...)
			outbt = append(outbt, vallenbt...)
			outbt = append(outbt, value...)
			if len(outbt) >= 4*1024*1024 {
				ff.Write(outbt)
				outbt = outbt[:0]
			}
		}
		if len(outbt) > 0 {
			ff.Write(outbt)
			outbt = outbt[:0]
		}
		ff.Close()
	}
}

func MakeDir(path string) bool {
	return MakePathDirExists(path)
}

func MakePathDirExists(path string) bool {
	path = ToAbsolutePath(path)
	dirnames := regexp.MustCompile("[\\\\/]+").Split(path, -1)
	if dirnames[len(dirnames)-1] != "" {
		dirnames = dirnames[:len(dirnames)-1]
	}
	ppath := ""
	for i := 0; i < len(dirnames); i++ {
		if ppath == "" {
			ppath += dirnames[i]
		} else {
			ppath += "/" + dirnames[i]
		}
		os.Mkdir(ppath, 0666)
	}
	return true
}

func DirGetAllPath(dir string, files *[]string) error {
	fis, fiser := os.ReadDir(dir)
	if fiser != nil {
		return fiser
	}
	for _, fi := range fis {
		if fi.IsDir() {
			*files = append(*files, dir+"/"+fi.Name()+"/")
			er := DirGetAllPath(dir+"/"+fi.Name(), files)
			if er != nil {
				return er
			}
		} else {
			*files = append(*files, dir+"/"+fi.Name())
		}
	}
	return nil
}

func d2t(d float64) string {
	return fmt.Sprintf("%0.3f", d)
}

//压缩byte,Compress Level, -1   (Default) 0x78,0x9C Compress Level, 0 0x78,0x1 Compress Level, 1 0x78,0x1 Compress Level, 2 0x78,0x5E Compress Level, 3 0x78,0x5E Compress Level, 4 0x78,0x5E Compress Level, 5 0x78,0x5E Compress Level, 6 0x78,0x9C Compress Level, 7 0x78,0xDA Compress Level, 8 0x78,0xDA Compress Level, 9 (Max compressLevel)0x78,0xDA
func ZipEncode(outbuf, input []byte, level int) []byte {
	var outbt []byte
	if outbuf != nil {
		outbt = outbuf
	}
	buf := bytes.NewBuffer(outbt)
	buf.Reset()
	compressor, err := zlib.NewWriterLevel(buf, level)
	if err != nil {
		panic("compress error 2362.")
	}
	compressor.Write(input)
	compressor.Close()
	if outbuf != nil {
		return outbt[:buf.Len()]
	} else {
		return buf.Bytes()
	}
}

//解压byte
func ZipDecode(outbuf, input []byte) []byte {
	if len(input) == 0 {
		return []byte{}
	}
	b := bytes.NewReader(input)
	r, err := zlib.NewReader(b)
	if r == nil {
		return make([]byte, 0)
	}
	defer r.Close()
	if err != nil {
		panic("unZipByte error 2394.")
	}
	b2 := bytes.NewBuffer(outbuf)
	b2.Reset()
	outw := bufio.NewWriter(b2)
	if _, err := io.Copy(outw, r); err != nil {
		panic("DelateDecode Error.")
	}
	if outbuf != nil {
		return outbuf[:b2.Len()]
	} else {
		return b2.Bytes()
	}
}
