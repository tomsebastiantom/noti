package bytebase

import (
    "os/exec"
    "log"
)

func Migrate(dsn, file string) error {
    cmd := exec.Command("bb", "migrate", "--dsn", dsn, "--file", file)
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Migration failed: %s", err.Error())
        return err
    }

    log.Printf("Migration succeeded: %s", string(output))
    return nil
}
