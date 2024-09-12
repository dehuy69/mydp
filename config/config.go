package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

// Config struct chứa cấu hình đường dẫn cho SQLite, Badger, và Parquet
type Config struct {
	DataFolderDefault string `mapstructure:"data_folder_default" envconfig:"DATA_FOLDER_DEFAULT"`
}

// LoadConfig tải cấu hình từ file YAML và biến môi trường
func LoadConfig() *Config {
	// Đọc cấu hình từ file YAML
	viper.SetConfigName("config_local") // Tên file cấu hình (không bao gồm đuôi mở rộng)
	viper.SetConfigType("yaml")         // Định dạng file cấu hình
	viper.AddConfigPath("config")       // Thư mục chứa file cấu hình

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		return nil
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Error unmarshalling config: %s\n", err)
		return nil
	}

	// Ghi đè với biến môi trường (nếu có)
	err = envconfig.Process("", &config)
	if err != nil {
		fmt.Printf("Error processing environment variables: %s\n", err)
		return nil
	}

	// Kiểm tra cấu hình đã tải
	fmt.Println("Data folder:", config.DataFolderDefault)

	return &config
}
