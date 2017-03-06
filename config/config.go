/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

// PeerConfig ...
type PeerConfig struct {
	Host      string
	Port      string
	EventHost string
	EventPort string
}

var myViper = viper.New()
var log = logging.MustGetLogger("fabric_sdk_go")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}`,
)

// InitConfig ...
// initConfig reads in config file
func InitConfig(configFile string) error {

	if configFile != "" {
		// create new viper
		myViper.SetConfigFile(configFile)
		// If a config file is found, read it in.
		err := myViper.ReadInConfig()

		if err == nil {
			log.Infof("Using config file: %s", myViper.ConfigFileUsed())
		} else {
			return fmt.Errorf("Fatal error config file: %v", err)
		}
	}

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	loggingLevelString := myViper.GetString("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != "" {
		log.Infof("fabric_sdk_go Logging level: %v", loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}
	logging.SetBackend(backendFormatter).SetLevel(logging.Level(logLevel), "fabric_sdk_go")

	return nil
}

// GetFabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func GetFabricClientViper() *viper.Viper {
	return myViper
}

// GetPeersConfig ...
func GetPeersConfig() []PeerConfig {
	peersConfig := []PeerConfig{}
	peers := myViper.GetStringMap("client.peers")
	for key, value := range peers {
		mm, ok := value.(map[string]interface{})
		var host string
		var port int
		var eventHost string
		var eventPort int

		if ok {
			host, _ = mm["host"].(string)
			port, _ = mm["port"].(int)
			eventHost, _ = mm["event_host"].(string)
			eventPort, _ = mm["event_port"].(int)
		} else {
			mm1 := value.(map[interface{}]interface{})
			host, _ = mm1["host"].(string)
			port, _ = mm1["port"].(int)
			eventHost, _ = mm1["event_host"].(string)
			eventPort, _ = mm1["event_port"].(int)
		}
		p := PeerConfig{Host: host, Port: strconv.Itoa(port), EventHost: eventHost, EventPort: strconv.Itoa(eventPort)}
		if p.Host == "" {
			panic(fmt.Sprintf("host key not exist or empty for %s", key))
		}
		if p.Port == "" {
			panic(fmt.Sprintf("port key not exist or empty for %s", key))
		}
		if p.EventHost == "" {
			panic(fmt.Sprintf("event_host not exist or empty for %s", key))
		}
		if p.EventPort == "" {
			panic(fmt.Sprintf("event_port not exist or empty for %s", key))
		}
		peersConfig = append(peersConfig, p)
	}
	return peersConfig

}

// IsTLSEnabled ...
func IsTLSEnabled() bool {
	return myViper.GetBool("client.tls.enabled")
}

// GetTLSCACertPool ...
func GetTLSCACertPool() *x509.CertPool {
	certPool := x509.NewCertPool()
	if myViper.GetString("client.tls.certificate") != "" {
		rawData, err := ioutil.ReadFile(myViper.GetString("tls.certificate"))
		if err != nil {
			panic(err)
		}
		certPool.AddCert(loadCAKey(rawData))
	}
	return certPool
}

// GetTLSServerHostOverride ...
func GetTLSServerHostOverride() string {
	return myViper.GetString("client.tls.serverhostoverride")
}

// IsSecurityEnabled ...
func IsSecurityEnabled() bool {
	return myViper.GetBool("client.security.enabled")
}

// TcertBatchSize ...
func TcertBatchSize() int {
	return myViper.GetInt("client.tcert.batch.size")
}

// GetSecurityAlgorithm ...
func GetSecurityAlgorithm() string {
	return myViper.GetString("client.security.hashAlgorithm")
}

// GetSecurityLevel ...
func GetSecurityLevel() int {
	return myViper.GetInt("client.security.level")

}

// GetOrdererHost ...
func GetOrdererHost() string {
	return myViper.GetString("client.orderer.host")
}

// GetMspID ...
func GetMspID() string {
	return myViper.GetString("client.msp.id")
}

// GetMspClientPath ...
func GetMspClientPath() string {
	return myViper.GetString("client.msp.clientPath")
}

// GetKeyStorePath ...
func GetKeyStorePath() string {
	return myViper.GetString("client.keystore.path")
}

// GetOrdererPort ...
func GetOrdererPort() string {
	return strconv.Itoa(myViper.GetInt("client.orderer.port"))
}

// loadCAKey
func loadCAKey(rawData []byte) *x509.Certificate {
	block, _ := pem.Decode(rawData)

	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return pub
}
