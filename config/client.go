package config

import (
	"fmt"
	rhttp "github.com/hashicorp/go-retryablehttp"
	"net/http"
	"strings"
	"time"
)

// policies
const (
	Empty             = ""
	DefaultPolicy     = "default"
	ExponentialPolicy = "exponential"
	JitterPolicy      = "jitter"
	ConstantPolicy    = "const"
)

func ConstantBackoff(min, _ time.Duration, _ int, _ *http.Response) time.Duration {
	return min
}

var policies = map[string]rhttp.Backoff{
	Empty:             rhttp.DefaultBackoff,
	DefaultPolicy:     rhttp.DefaultBackoff,
	ExponentialPolicy: rhttp.DefaultBackoff,
	JitterPolicy:      rhttp.LinearJitterBackoff,
	ConstantPolicy:    ConstantBackoff,
}

type RetryConfig struct {
	WaitMin     time.Duration `yaml:"wait_min"`
	WaitMax     time.Duration `yaml:"wait_max"`
	MaxAttempts int           `yaml:"max_attempts"`
	Policy      string        `yaml:"policy,omitempty"`
	backoff     rhttp.Backoff `yaml:"-"`
	isValid     bool          `yaml:"-"`
}

// Validate must be called once, after rc has been constructed / unmarshalled
func (rc *RetryConfig) Validate() (err error) {
	if rc != nil {
		if err = validDurations(0, rc.WaitMin, false); err == nil {
			if err = validDurations(rc.WaitMin, rc.WaitMin, true); err == nil {
				if err = validPositive(rc.MaxAttempts); err == nil {
					if rc.backoff = policies[strings.ToLower(rc.Policy)]; rc.backoff == nil {
						err = fmt.Errorf("invalid backoff policy %s", rc.Policy)
					}
				}
			}
		}
		rc.isValid = err == nil
	}
	return
}

// NewClient should be called only after Validate has been called, to make sure
// that rc is a valid RetryConfig
func (rc *RetryConfig) NewClient(rt http.RoundTripper, logger interface{}) (*http.Client, error) {
	c := rhttp.NewClient()
	if rc != nil {
		if !rc.isValid {
			return nil, fmt.Errorf("retry configuration is not valid")
		}
		c.RetryWaitMin = rc.WaitMin
		c.RetryWaitMax = rc.WaitMax
		c.RetryMax = rc.MaxAttempts
		c.Backoff = rc.backoff
	}
	c.HTTPClient = &http.Client{Transport: rt}
	// set the logger (rhttp default logger is debug-level, too verbose)
	if logger != nil {
		switch logger.(type) {
		case rhttp.Logger, rhttp.LeveledLogger:
			// OK
		default:
			return nil, fmt.Errorf("invalid logger type %T", logger)
		}
	}
	c.Logger = logger
	return c.StandardClient(), nil
}

func validDurations(d1, d2 time.Duration, equalAllowed bool) (err error) {
	var test bool
	var operator string
	if equalAllowed {
		test = d2 >= d1
		operator = ">"
	} else {
		test = d2 > d1
		operator = ">="
	}
	if !test {
		err = fmt.Errorf("invalid durations: %v %s %v", d1, operator, d2)
	}
	return
}

func validPositive(n int) (err error) {
	if n <= 0 {
		err = fmt.Errorf("number %d must be positive", n)
	}
	return
}
