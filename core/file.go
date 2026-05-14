package core

import (
	"bytes"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/pelletier/go-toml"
)

const pemPkHeader = "PRIVATE KEY for "

// MaxTomlMapFileSize bounds how much LoadTomlFileToMap is willing to
// allocate from disk. ISSUE-046: previously the function did
// `make([]byte, fileinfo.Size())` with no cap, so a malicious or
// corrupted config (toml/gas-schedule/...) could request gigabytes of
// allocation and OOM the process at startup. 16 MiB is generous
// enough for every legitimate config file in this codebase (gas
// schedules, node configs, p2p configs are all well under 1 MB) and
// restrictive enough to make accidental or hostile oversize obvious
// at config-load time.
const MaxTomlMapFileSize = 16 * 1024 * 1024

// ArgCreateFileArgument will hold the arguments for a new file creation method call
type ArgCreateFileArgument struct {
	Directory     string
	Prefix        string
	FileExtension string
}

// OpenFile method opens the file from given path - does not close the file
func OpenFile(relativePath string) (*os.File, error) {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// LoadTomlFile method to open and decode toml file
func LoadTomlFile(dest interface{}, relativePath string) error {
	f, err := OpenFile(relativePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return toml.NewDecoder(f).Decode(dest)
}

// SaveTomlFile will open and save data to toml file
func SaveTomlFile(src interface{}, relativePath string) error {
	f, err := os.Create(relativePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return toml.NewEncoder(f).Encode(src)
}

// LoadTomlFileToMap opens and decodes a toml file as a map[string]interface{}
func LoadTomlFileToMap(relativePath string) (map[string]interface{}, error) {
	f, err := OpenFile(relativePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	buffer, err := readBoundedTomlMapFile(relativePath, f)
	if err != nil {
		return nil, err
	}

	loadedTree, err := toml.Load(string(buffer))
	if err != nil {
		return nil, err
	}

	loadedMap := loadedTree.ToMap()

	return loadedMap, nil
}

func readBoundedTomlMapFile(relativePath string, f *os.File) ([]byte, error) {
	fileinfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return readBoundedTomlMap(relativePath, fileinfo.Size(), f)
}

func readBoundedTomlMap(relativePath string, fileSize int64, reader io.Reader) ([]byte, error) {
	// ISSUE-046: bound the allocation. Reject the read up-front if the
	// file size exceeds MaxTomlMapFileSize so callers see a clear
	// "config too large" error rather than running out of memory at
	// `make([]byte, filesize)`.
	if fileSize < 0 || fileSize > MaxTomlMapFileSize {
		return nil, fmt.Errorf("toml file %q size %d exceeds the maximum allowed %d bytes",
			relativePath, fileSize, MaxTomlMapFileSize)
	}

	// Read one byte past the configured limit so a file that grows after
	// Stat is rejected instead of silently truncated before TOML parsing.
	buffer, err := io.ReadAll(io.LimitReader(reader, MaxTomlMapFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(buffer) > MaxTomlMapFileSize {
		return nil, fmt.Errorf("toml file %q grew beyond the maximum allowed %d bytes while reading",
			relativePath, MaxTomlMapFileSize)
	}

	return buffer, nil
}

// LoadJsonFile method to open and decode json file
func LoadJsonFile(dest interface{}, relativePath string) error {
	f, err := OpenFile(relativePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return json.NewDecoder(f).Decode(dest)
}

// CreateFile opens or creates a file relative to the default path
func CreateFile(arg ArgCreateFileArgument) (*os.File, error) {
	absPath, err := filepath.Abs(arg.Directory)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(absPath, 0700)
	if err != nil {
		return nil, err
	}

	fileName := time.Now().Format("2006-01-02T15-04-05")
	fileName = strings.Replace(fileName, "T", "-", 1)
	if arg.Prefix != "" {
		fileName = arg.Prefix + "-" + fileName
	}

	return os.OpenFile(
		filepath.Join(absPath, fileName+"."+arg.FileExtension),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		FileModeUserReadWrite)
}

// LoadSkPkFromPemFile loads the secret key and existing public key bytes stored in the file
func LoadSkPkFromPemFile(relativePath string, skIndex int) ([]byte, string, error) {
	if skIndex < 0 {
		return nil, "", ErrInvalidIndex
	}

	file, err := OpenFile(relativePath)
	if err != nil {
		return nil, "", err
	}

	defer func() {
		_ = file.Close()
	}()

	buff, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("%w while reading %s file", err, relativePath)
	}
	if len(buff) == 0 {
		return nil, "", fmt.Errorf("%w while reading %s file", ErrEmptyFile, relativePath)
	}

	var blkRecovered *pem.Block

	for i := 0; i <= skIndex; i++ {
		if len(buff) == 0 {
			//less private keys present in the file than required
			return nil, "", fmt.Errorf("%w while reading %s file, invalid index %d", ErrInvalidIndex, relativePath, i)
		}

		blkRecovered, buff = pem.Decode(buff)
		if blkRecovered == nil {
			return nil, "", fmt.Errorf("%w while reading %s file, error decoding", ErrPemFileIsInvalid, relativePath)
		}
	}

	if blkRecovered == nil {
		return nil, "", ErrNilPemBLock
	}

	blockType := blkRecovered.Type

	if !strings.HasPrefix(blockType, pemPkHeader) {
		return nil, "", fmt.Errorf("%w missing '%s' in block type", ErrPemFileIsInvalid, pemPkHeader)
	}

	blockTypeString := blockType[len(pemPkHeader):]
	if !isValidPemPublicKeySuffix(blockTypeString) {
		return nil, "", fmt.Errorf("%w invalid public key suffix in block type", ErrPemFileIsInvalid)
	}

	return blkRecovered.Bytes, blockTypeString, nil
}

// LoadAllKeysFromPemFile loads all the secret keys and existing public key bytes stored in the file
func LoadAllKeysFromPemFile(relativePath string) ([][]byte, []string, error) {
	file, err := OpenFile(relativePath)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	buff, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("%w while reading %s file", err, relativePath)
	}
	if len(buff) == 0 {
		return nil, nil, fmt.Errorf("%w while reading %s file", ErrEmptyFile, relativePath)
	}

	var blkRecovered *pem.Block
	privateKeys := make([][]byte, 0)
	publicKeys := make([]string, 0)

	for {
		if len(buff) == 0 {
			break
		}

		blkRecovered, buff = pem.Decode(buff)
		if blkRecovered == nil {
			return nil, nil, fmt.Errorf("%w while reading %s file, error decoding", ErrPemFileIsInvalid, relativePath)
		}
		buff = bytes.TrimSpace(buff)

		blockType := blkRecovered.Type
		if !strings.HasPrefix(blockType, pemPkHeader) {
			return nil, nil, fmt.Errorf("%w missing '%s' in block type", ErrPemFileIsInvalid, pemPkHeader)
		}

		blockTypeString := blockType[len(pemPkHeader):]
		if !isValidPemPublicKeySuffix(blockTypeString) {
			return nil, nil, fmt.Errorf("%w invalid public key suffix in block type", ErrPemFileIsInvalid)
		}

		privateKeys = append(privateKeys, blkRecovered.Bytes)
		publicKeys = append(publicKeys, blockTypeString)
	}

	return privateKeys, publicKeys, nil
}

func isValidPemPublicKeySuffix(suffix string) bool {
	if suffix == "" || strings.TrimSpace(suffix) != suffix {
		return false
	}
	for _, char := range suffix {
		if unicode.IsControl(char) {
			return false
		}
	}

	return true
}

// SaveSkToPemFile saves secret key bytes in the file
func SaveSkToPemFile(file *os.File, identifier string, skBytes []byte) error {
	if file == nil {
		return ErrNilFile
	}
	if !isValidPemPublicKeySuffix(identifier) {
		return fmt.Errorf("%w invalid public key suffix in block type", ErrPemFileIsInvalid)
	}

	blk := pem.Block{
		Type:  "PRIVATE KEY for " + identifier,
		Bytes: skBytes,
	}

	return pem.Encode(file, &blk)
}
