package entity

type Config struct {
	MockSize   int    `env:"MOCK_SIZE" default:"1000"`
	OutputDir  string `env:"OUTPUT_DIR" default:"./output"`
	OutputFile string `env:"OUTPUT_FILE" default:"benchmark.json"`
}
