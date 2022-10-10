package models

import "fmt"

type Config struct {
	Port     string `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Request struct {
	LowPrice  int64 `json:"low_price,required"`
	HighPrice int64 `json:"high_price,required"`
}
type Response struct {
	Notebook []Notebook `json:"goods"`
}
type Notebook struct {
	Name             string  `json:"name,required"`
	Price            int     `json:"price,required"`
	Ref              string  `json:"reference,required"`
	ScreenResolution string  `json:"screen_resolution,required"`
	CPUFrequency     float64 `json:"cpu_frequency,required"`
	CPUCores         int     `json:"cpu_cores,required"`
	RAM              int     `json:"ram,required"`
	GPURAM           int     `json:"gpuram,required"`
	Storage          int     `json:"storage,required"`
	Rating           float64 `json:"rating,required"`
}

func (n Notebook) String() string {
	return fmt.Sprintf("{\n%v\n%v\n%v\n}", n.Name, n.Price, n.Ref)
}
