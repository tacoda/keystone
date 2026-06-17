package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// snapshotCmd groups save/list/restore. Snapshots are tarballs of
// `.keystone/` stashed under `.keystone-snapshots/` at the project
// root. Cheap insurance before `keystone migrate`, destructive
// prune passes, or experimental policy installs.
func snapshotCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "snapshot",
		Short: "Save / list / restore local snapshots of .keystone/",
	}
	c.AddCommand(snapshotSaveCmd())
	c.AddCommand(snapshotListCmd())
	c.AddCommand(snapshotRestoreCmd())
	return c
}

func snapshotsDir(projectDir string) string {
	return filepath.Join(projectDir, ".keystone-snapshots")
}

func snapshotSaveCmd() *cobra.Command {
	var (
		dir   string
		label string
	)
	c := &cobra.Command{
		Use:   "save",
		Short: "Snapshot .keystone/ to .keystone-snapshots/<ts>-<label>.tar.gz",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			ks := filepath.Join(abs, ".keystone")
			if _, err := os.Stat(ks); err != nil {
				return fmt.Errorf(".keystone/ not found at %s — nothing to snapshot", abs)
			}
			snapDir := snapshotsDir(abs)
			if err := os.MkdirAll(snapDir, 0o755); err != nil {
				return err
			}
			ts := time.Now().UTC().Format("20060102-150405")
			name := ts
			if label != "" {
				name = ts + "-" + sanitizeLabel(label)
			}
			out := filepath.Join(snapDir, name+".tar.gz")
			if err := tarGz(ks, out, abs); err != nil {
				return err
			}
			info, _ := os.Stat(out)
			fmt.Fprintf(os.Stdout, "✓ saved %s (%d bytes)\n", out, info.Size())
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	c.Flags().StringVar(&label, "label", "", "Optional label appended to the filename.")
	return c
}

func snapshotListCmd() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "list",
		Short: "List local snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			entries, err := os.ReadDir(snapshotsDir(abs))
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			names := []string{}
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".tar.gz") {
					names = append(names, e.Name())
				}
			}
			sort.Strings(names)
			for _, n := range names {
				fmt.Println(n)
			}
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	return c
}

func snapshotRestoreCmd() *cobra.Command {
	var dir string
	c := &cobra.Command{
		Use:   "restore <snapshot>",
		Short: "Restore .keystone/ from a snapshot — destructive, asks for confirm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			name := args[0]
			if !strings.HasSuffix(name, ".tar.gz") {
				name += ".tar.gz"
			}
			src := filepath.Join(snapshotsDir(abs), name)
			if _, err := os.Stat(src); err != nil {
				return fmt.Errorf("snapshot %s not found", src)
			}
			ks := filepath.Join(abs, ".keystone")
			fmt.Fprintf(os.Stderr, "⚠ this will REMOVE %s and restore from %s. Type 'yes' to proceed: ", ks, src)
			var ack string
			_, _ = fmt.Scanln(&ack)
			if strings.TrimSpace(ack) != "yes" {
				return fmt.Errorf("cancelled")
			}
			if err := os.RemoveAll(ks); err != nil {
				return err
			}
			if err := untarGz(src, abs); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "✓ restored from %s\n", src)
			return nil
		},
	}
	c.Flags().StringVar(&dir, "dir", ".", "Project root (defaults to cwd).")
	return c
}

func sanitizeLabel(s string) string {
	out := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
			return r
		case r == ' ':
			return '-'
		}
		return -1
	}, s)
	if len(out) > 40 {
		out = out[:40]
	}
	return out
}

func tarGz(srcAbs, dstAbs, projectAbs string) error {
	f, err := os.Create(dstAbs)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	return filepath.Walk(srcAbs, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(projectAbs, p)
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			in, err := os.Open(p)
			if err != nil {
				return err
			}
			_, err = io.Copy(tw, in)
			in.Close()
			return err
		}
		return nil
	})
}

func untarGz(src, projectAbs string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		out := filepath.Join(projectAbs, hdr.Name)
		if !strings.HasPrefix(out, projectAbs) {
			return fmt.Errorf("snapshot entry %q escapes project dir", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(out, os.FileMode(hdr.Mode)|0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			fout, err := os.Create(out)
			if err != nil {
				return err
			}
			if _, err := io.Copy(fout, tr); err != nil {
				fout.Close()
				return err
			}
			fout.Close()
		}
	}
}
