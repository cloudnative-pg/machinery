package wals

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cloudnative-pg/machinery/pkg/log"
)

// WALList is a structure that contains the list of WAL files that are ready to be archived
type WALList struct {
	pgDataPath string
	// points to a wal, example of content: "pg_wal/000000010000000000000001"
	Ready []string
	// points to a wal, example of content: "pg_wal/000000010000000000000001"
	Done           []string
	HasMoreResults bool
}

// RemoveReadyItem removes a WAL file from the list of ready WAL files
func (w *WALList) RemoveReadyItem(walName string) {
	var filtered []string
	for _, wal := range w.Ready {
		if !strings.HasSuffix(wal, walName) {
			filtered = append(filtered, wal)
		}
	}
	w.Ready = filtered
}

// ReadyItemsToSlice returns the list of ready WAL files as a slice
func (w *WALList) ReadyItemsToSlice() []string {
	return slices.Clone(w.Ready)
}

// MarkAsDone moves a WAL file from the list of ready WAL files to the list of done WAL files
func (w *WALList) MarkAsDone(ctx context.Context, walName string) error {
	contextLogger := log.FromContext(ctx)
	// Extract the base name of the walName to ensure consistency
	walBaseName := filepath.Base(walName)

	readyPath := path.Join(
		getArchiveStatusPath(w.pgDataPath),
		walBaseName+".ready",
	)
	donePath := path.Join(
		getArchiveStatusPath(w.pgDataPath),
		walBaseName+".done",
	)

	err := os.Rename(readyPath, donePath)
	if err != nil {
		contextLogger.Error(
			err,
			"failed to rename WAL file",
			"readyPath", readyPath,
			"donePath", donePath,
		)
		return err
	}

	w.RemoveReadyItem(walName)
	w.Done = append(w.Done, walName)
	return nil
}

// GatherReadyWALFilesConfig is the configuration for GatherReadyWALFiles
type GatherReadyWALFilesConfig struct {
	PgDataPath string
	MaxResults int
	SkipWALs   []string
}

func (c GatherReadyWALFilesConfig) getPgDataPath() string {
	if c.PgDataPath == "" {
		return os.Getenv("PGDATA")
	}
	return c.PgDataPath
}

func (c GatherReadyWALFilesConfig) shouldSkipWAL(walPath string) bool {
	for _, walToSkip := range c.SkipWALs {
		if strings.HasSuffix(walPath, walToSkip) {
			return true
		}
	}
	return false
}

// GatherReadyWALFiles reads from the archived status the list of WAL files
// that can be archived.
func GatherReadyWALFiles(
	ctx context.Context,
	config GatherReadyWALFilesConfig,
) *WALList {
	contextLog := log.FromContext(ctx)
	archiveStatusPath := getArchiveStatusPath(config.getPgDataPath())
	noMoreWALFilesNeeded := errors.New("no more files needed")

	var walList []string
	err := filepath.WalkDir(archiveStatusPath, func(path string, d os.DirEntry, err error) error {
		// If err is set, it means the current path is a directory and the readdir raised an error
		// The only available option here is to skip the path and log the error.
		if err != nil {
			contextLog.Error(err, "failed reading path", "path", path)
			return filepath.SkipDir
		}

		// We don't process directories beside the archive status path
		if d.IsDir() {
			// We want to proceed exploring the archive status folder
			if path == archiveStatusPath {
				return nil
			}

			return filepath.SkipDir
		}

		// We only process ready files
		if !strings.HasSuffix(path, ".ready") {
			return nil
		}

		if len(walList) >= config.MaxResults {
			return noMoreWALFilesNeeded
		}

		// We are already archiving the requested WAL file,
		// and we need to avoid archiving it twice.
		// requestedWALFile is usually "pg_wal/wal_file_name" and
		// we compare it with the path we read
		if config.shouldSkipWAL(path) {
			return nil
		}

		walFileName := strings.TrimSuffix(filepath.Base(path), ".ready")

		walList = append(
			walList,
			filepath.Join(config.getPgDataPath(), "pg_wal", walFileName),
		)
		return nil
	})

	// In this point err must be nil or noMoreWALFilesNeeded, if it is something different
	// there is a programming error
	if err != nil && !errors.Is(err, noMoreWALFilesNeeded) {
		contextLog.Error(err, "unexpected error while reading the list of WAL files to archive")
	}

	return &WALList{
		Ready:          walList,
		HasMoreResults: errors.Is(err, noMoreWALFilesNeeded),
		pgDataPath:     config.getPgDataPath(),
	}
}

func getArchiveStatusPath(pgDataPath string) string {
	pgWalDirectory := path.Join(pgDataPath, "pg_wal")
	archiveStatusPath := path.Join(pgWalDirectory, "archive_status")
	return archiveStatusPath
}
