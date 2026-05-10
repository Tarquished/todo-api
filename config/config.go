package config

import "github.com/spf13/viper"

func LoadConfig() error {
	viper.SetConfigFile(".env")

	// INI DIA KUNCI JAWABANNYA! 👇
	// Harus dipanggil supaya Viper baca variabel dari Docker/Railway
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	// Kita abaikan error kalau file .env nggak ada, karena di Docker memang nggak ada
	if err != nil {
		return err
	}

	return nil
}
