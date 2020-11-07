package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oxzi/pmwiki-pagefileformat-go"
)

// pmWikiDir and gitDir are the directories to be used for PmWiki input and git output.
var pmWikiDir, gitDir string

// init handles the setup; flag parsing and the like.
func init() {
	flag.StringVar(&pmWikiDir, "pmwiki", "", "path of PmWiki's wiki.d directory")
	flag.StringVar(&gitDir, "git", "", "path to the output git repository")

	flag.Parse()

	if pmWikiDir == "" || gitDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	for _, dir := range []string{pmWikiDir, gitDir} {
		if stat, err := os.Stat(dir); os.IsNotExist(err) {
			log.WithField("directory", dir).Fatal("Directory does not exist")
		} else if !stat.IsDir() {
			log.WithField("directory", dir).Fatal("Directory is not a directory")
		}
	}
}

// pmWikiRevisions returns all successfully parsed revisions of a PmWiki.
func pmWikiRevisions() (revs []pmwiki.PageFile) {
	isDeleted := regexp.MustCompile(`.*,del-(\d+)$`)

	fs, err := ioutil.ReadDir(pmWikiDir)
	if err != nil {
		log.WithField("pmwiki", pmWikiDir).WithError(err).Fatal("Cannot read PmWiki directory")
	}

	for _, f := range fs {
		if f.Name()[0] == '.' {
			continue
		}

		logger := log.WithField("file", f.Name())

		r, err := os.Open(path.Join(pmWikiDir, f.Name()))
		if err != nil {
			logger.WithError(err).Fatal("Cannot open file")
		}

		pageFile, err := pmwiki.ParsePageFile(r)
		if err != nil {
			logger.WithError(err).Error("Cannot parse page file")
			goto fail
		}

		if matches := isDeleted.FindStringSubmatch(f.Name()); len(matches) > 0 {
			if unixInt, err := strconv.ParseInt(matches[1], 10, 64); err != nil {
				logger.WithError(err).Error("Cannot parse deletion timestamp")
				goto fail
			} else {
				pageFile.Deleted = time.Unix(unixInt, 0).UTC()
			}
		}

		if err := pageFile.Revisions(func(pf pmwiki.PageFile) { revs = append(revs, pf) }); err != nil {
			logger.WithError(err).Error("Cannot parse page file revisions")
		}

	fail:
		err = r.Close()
		if err != nil {
			logger.WithError(err).Fatal("Cannot close file")
		}
	}

	return
}

// createCommit for this PageFile revision within the git repository.
func createCommit(revision pmwiki.PageFile) error {
	basePath, err := filepath.Abs(gitDir)
	if err != nil {
		return err
	}
	filename := path.Join(basePath, path.Clean(strings.Replace(revision.Name, ".", "/", 1)))

	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		return fmt.Errorf("mkdir errored, %w", err)
	}

	author := revision.Author
	if author == "" {
		author = "pmwiki"
	}
	message := ""

	if revision.Text != "" {
		// Create or alter the file
		if f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			return err
		} else if _, err := f.WriteString(revision.Text); err != nil {
			return err
		} else if err := f.Close(); err != nil {
			return err
		}

		addCmd := exec.Command("git", "-C", gitDir, "add", filename)
		if out, err := addCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add errored: %w, %s", err, out)
		}

		message = fmt.Sprintf("Modified %s", revision.Name)
	} else {
		// Delete the file
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.WithFields(log.Fields{
				"file":     revision.Name,
				"revision": revision.Time,
			}).Warn("Revision has empty text, but does not exists")
			return nil
		}

		rmCmd := exec.Command("git", "-C", gitDir, "rm", filename)
		if out, err := rmCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git rm errored: %w, %s", err, out)
		}

		message = fmt.Sprintf("Deleted %s", revision.Name)
	}

	statusCmd := exec.Command("git", "-C", gitDir, "status", "--porcelain")
	if out, err := statusCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git status errored: %w, %s", err, out)
	} else if strings.TrimSpace(string(out)) == "" {
		log.WithFields(log.Fields{
			"file":     revision.Name,
			"revision": revision.Time,
		}).Warn("Revision did not produced any `git commit`able difference")
		return nil
	}

	commitCmd := exec.Command("git", "-C", gitDir, "commit", "-m", message)
	commitCmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", author),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s@pmwiki", author),
		fmt.Sprintf("GIT_AUTHOR_DATE=%s", revision.Time.Format(time.RFC1123Z)),
		fmt.Sprintf("GIT_COMMITTER_NAME=%s", author),
		fmt.Sprintf("GIT_COMMITTER_EMAIL=%s@pmwiki", author),
		fmt.Sprintf("GIT_COMMITTER_DATE=%s", revision.Time.Format(time.RFC1123Z)))
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit errored: %w, %s", err, out)
	}

	return nil
}

func main() {
	log.WithFields(log.Fields{"pmwiki": pmWikiDir, "git": gitDir}).Info("Starting PmWiki to git conversion")

	revs := pmWikiRevisions()
	log.WithField("revisions", len(revs)).Info("Finished parsing revisions")

	sort.Sort(pmwiki.ByTime(revs))

	for _, rev := range revs {
		if err := createCommit(rev); err != nil {
			log.WithField("file", rev.Name).WithError(err).Fatal("Creating git commit failed")
		}
	}

	log.WithField("git", gitDir).Info("Finished git import")
}
