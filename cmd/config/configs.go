package configs

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigInfo struct {
	Network
	Blockchain
}

type Network struct {
	ProtocolType     string
	RPCListenAddress string
	ClientAPIAddress string
	P2PListenAddress string
	P2PBootstraps    []string
	ProtocolVersion  uint32
	NetworkID        uint32
	ServerCrt        string
	ServerKey        string
}

type Blockchain struct {
	BlockDbPath    string
	KeyStoragePath string
}

var Configinfos ConfigInfo

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		// logrus.Infof("block insert to db by hash: %s", block.Hash.Hex())
		logrus.Infof("Fatal error config file: %s \n", err)
	}

	Configinfos.Network.ProtocolType = viper.GetString("network.protocol_type")
	Configinfos.Network.RPCListenAddress = viper.GetString("network.RPCListenAddress")
	Configinfos.Network.ClientAPIAddress = viper.GetString("network.ClientAPIAddress")
	Configinfos.Network.P2PListenAddress = viper.GetString("network.P2PListenAddress")
	Configinfos.Network.ProtocolVersion = viper.GetUint32("network.ProtocolVersion")
	Configinfos.Network.P2PBootstraps = viper.GetStringSlice("network.P2PBootstraps")
	Configinfos.Network.NetworkID = viper.GetUint32("network.NetworkID")
	Configinfos.Network.ServerCrt = viper.GetString("network.server_crt")
	Configinfos.Network.ServerCrt = viper.GetString("network.server_key")

	Configinfos.Blockchain.BlockDbPath = viper.GetString("blockchain.BlockDbPath")
	Configinfos.Blockchain.KeyStoragePath = viper.GetString("blockchain.KeyStoragePath")
}
func GetConfig() ConfigInfo {
	return Configinfos
}
