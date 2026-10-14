package deployment

import (
 "fmt"
 "strings"
)

const (
 Cloud="cloud"
 SelfHosted="self_hosted"
)

type Config struct {
 Mode string `env:"DEPLOYMENT_MODE" envDefault:"cloud"`
 DeploymentID string `env:"DEPLOYMENT_ID"`
}

func (c Config) SelfHosted() bool {return c.Mode==SelfHosted}
func (c Config) Validate() error {
 mode:=strings.TrimSpace(c.Mode)
 if mode=="" {mode=Cloud}
 if mode!=Cloud && mode!=SelfHosted{return fmt.Errorf("unsupported DEPLOYMENT_MODE %q",c.Mode)}
 return nil
}
