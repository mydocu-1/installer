// "Installs" files and directories into the HOME dir.
// Files are "installed" as hardlinks.
// Directories are "installed" as symlinks.
package main

import (
  "flag"
  "fmt"
  "gopkg.in/yaml.v3"
  "io/ioutil"
  "log"
  "os"
  "path/filepath"
)

var (
  // Flags.
  fConfigFile = flag.String("config", "", "Config YAML file.")
)

// Installer pkg.

type Installer struct {
  srcDir string
  destDir string
  backupDir string
}

func (i *Installer) Install(src string, dest string) error {
  if err := i.backup(dest); err != nil {
    return fmt.Errorf("Failed to backup %s. Error: %w", dest, err)
  }

  if err := i.install(src, dest); err != nil {
    return fmt.Errorf("Failed to install %s to %s. Error: %w", src, dest, err)
  }
  return nil
}

func (i Installer) install(src string, dest string) error {
  src = filepath.Join(i.srcDir, src)
  dest = filepath.Join(i.destDir, dest)

  fi, err := os.Lstat(src)
  if err != nil {
   return err
  }
  if fi.IsDir() {
    return os.Symlink(src, dest)
  } else {
    return os.Link(src, dest)
  }
}

func (i *Installer) backup(dest string) error {
  src := filepath.Join(i.destDir, dest)
  if _, err := os.Lstat(src); err == nil {
    back, err := i.backupPath(dest)
    if err != nil {
      return err
    }
    if err := os.MkdirAll(filepath.Dir(back), 0755); err != nil {
      return err
    }
    if err := os.Rename(src, back); err != nil {
      return err
    }
  }
  return nil
}

func (i *Installer) backupPath(dest_name string) (string, error) {
  if i.backupDir == "" {
    tmp, err := os.MkdirTemp(i.destDir, "backup")
    if err != nil {
      log.Fatal(err)
      return "", err
    }
    i.backupDir = tmp
    log.Printf("Existing files will be preserved up in %s\n", i.backupDir)
  }
  return filepath.Join(i.backupDir, dest_name), nil
}


// Config pkg.

type configData struct {
  Entries map[string]string `yaml:"entries"`
}


// Main pkg.

func main(){
  flag.Parse()

  log.Printf("Reading config file %s\n", *fConfigFile)
  yaml_data, err := ioutil.ReadFile(*fConfigFile)
  if err != nil {
    log.Fatal(err)
    return
  }

  log.Println("Parsing config YAML...")
  data := configData{}
  if err := yaml.Unmarshal([]byte(yaml_data), &data); err != nil {
    log.Fatal(err)
    return
  }
  log.Printf("Config data loaded successfully: %d entries to be installed.\n", len(data.Entries))

  srcDir, err := filepath.Abs(filepath.Dir(*fConfigFile))
  if err != nil {
    log.Fatal(err)
    return
  }

  home, err := os.UserHomeDir()
  if err != nil {
    log.Fatal(err)
    return
  }

  log.Printf("Installing files from %s to %s\n", srcDir, home)
  i := &Installer{srcDir: srcDir, destDir: home}

  for src, dest := range data.Entries {
    i.Install(src, dest)
  }

  log.Println("Done!")
}

