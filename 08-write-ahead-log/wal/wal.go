package wal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type Record struct {
	Sequence uint64 `json:"sequence"`
	Payload  string `json:"payload"`
	Checksum string `json:"checksum"`
}

type Log struct {
	mu           sync.Mutex
	file         *os.File
	nextSequence uint64
}

func Open(path string) (*Log, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	lines := bytes.Split(data, []byte{'\n'})
	completeLines := lines

	// A final line without '\n' is an incomplete record.
	if len(lines) > 0 && len(lines[len(lines)-1]) > 0 {
		partialRecord := lines[len(lines)-1]
		completeLines = lines[:len(lines)-1]

		validBytes := len(data) - len(partialRecord)

		if err := file.Close(); err != nil {
			return nil, err
		}

		if err := os.Truncate(path, int64(validBytes)); err != nil {
			return nil, err
		}

		file, err = os.OpenFile(
			path,
			os.O_CREATE|os.O_RDWR|os.O_APPEND,
			0644,
		)
		if err != nil {
			return nil, err
		}
	}

	logFile := &Log{
		file: file,
	}

	for _, line := range completeLines {
		if len(line) == 0 {
			continue
		}

		var record Record

		if err := json.Unmarshal(line, &record); err != nil {
			file.Close()
			return nil, err
		}

		expectedChecksum := calculateChecksum(
			record.Sequence,
			record.Payload,
		)

		if record.Checksum != expectedChecksum {
			file.Close()

			return nil, fmt.Errorf(
				"invalid checksum for sequence %d",
				record.Sequence,
			)
		}

		if record.Sequence >= logFile.nextSequence {
			logFile.nextSequence = record.Sequence + 1
		}
	}

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		file.Close()
		return nil, err
	}

	return logFile, nil
}

func (l *Log) Append(payload string) (Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	record := Record{
		Sequence: l.nextSequence,
		Payload:  payload,
		Checksum: calculateChecksum(
			l.nextSequence,
			payload,
		),
	}

	data, err := json.Marshal(record)
	if err != nil {
		return Record{}, err
	}

	data = append(data, '\n')

	if _, err := l.file.Write(data); err != nil {
		return Record{}, err
	}

	if err := l.file.Sync(); err != nil {
		return Record{}, err
	}

	l.nextSequence++
	return record, nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return nil
	}

	err := l.file.Sync()
	if closeErr := l.file.Close(); err == nil {
		err = closeErr
	}

	l.file = nil
	return err
}

func calculateChecksum(sequence uint64, payload string) string {
	input := fmt.Sprintf("%d:%s", sequence, payload)

	sum := sha256.Sum224([]byte(input))

	return hex.EncodeToString(sum[:])
}
