package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type FundDataItem struct {
	Date string
	// 小数点后4位小数，转换为整数来进行计算
	DayProfit int64
	// 小数点后4位小数，转换为整数来计算
	TotalProfit int64
	ProfitRait  string
}

func main() {

	fh, error := os.Open("./data/efunds_sh50etf_110003.csv")
	if error != nil {
		fmt.Println(error.Error())
		os.Exit(-1)
	}
	full, err := ioutil.ReadAll(fh)
	if err != nil {
		fmt.Println(error.Error())
		os.Exit(-1)
	}
	fullStr := string(full)

	fLine := strings.Split(fullStr, "\n")
	fmt.Println(fLine[0])

	//获得了一行行的数据
	var fundData []FundDataItem

	for _, s := range fLine {

		it := strings.Fields(s)
		if len(it) > 3 {
			x, err := strconv.ParseFloat(it[1], 64)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			y, err := strconv.ParseFloat(it[2], 64)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			fdi := FundDataItem{

				Date:        it[0],
				DayProfit:   int64(x * 10000),
				TotalProfit: int64(y * 10000),
				ProfitRait:  it[3],
			}

			fmt.Printf(" the result of fundDataItem: %+v\n", fdi)
			fundData = append(fundData, fdi)

		}
	}

}