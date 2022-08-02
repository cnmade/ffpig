package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"text/tabwriter"

	"github.com/shopspring/decimal"
)

type AccountLog struct {
	Id             int64
	TradeId        int64
	Date           string
	AccountType    int
	Amount         int64
	AfterAmount    int64
	ProfitExchange float64
	Desc           string
}

type AccountBook struct {
	// 1. 支出账户，用于记录一共花多少钱
	CostAccount int64
	// 2. 收益账户，记录收益账户，一共赚多少钱
	EarnAccount int64
	// 3. 记录个人基金账户，一共有多少及价值
	FundAccount int64
	// 4. 记录基金公司账户的 金额
	PlatformAccount int64
}

type FundDataItem struct {
	Date string
	// 小数点后4位小数，转换为整数来进行计算
	DayProfit int64
	// 小数点后4位小数，转换为整数来计算
	TotalProfit int64
	ProfitRait  string
}
type BonousRatioPair struct {
	Buy  decimal.Decimal
	Sell decimal.Decimal
}

// 手续费是0.01
var feeRatio = decimal.NewFromFloat(0.0005)

//一万块钱开始
var startBalance = decimal.NewFromInt(10000 * 10000)

//ID发号器
var IdChain int64 = 0

var tradeIdSerial int64 = 0

// 1万
var ONEW = decimal.NewFromInt(10000)

//获得int64值
func AsInt64(i decimal.Decimal) int64 {
	f, _ := i.BigFloat().Int64()
	return f
}
func GetNextId() int64 {
	nextId := atomic.LoadInt64(&IdChain) + 1
	atomic.StoreInt64(&IdChain, nextId)
	return nextId
}

func GetNextTradeId() int64 {

	nextId := atomic.LoadInt64(&tradeIdSerial) + 1
	atomic.StoreInt64(&tradeIdSerial, nextId)
	return nextId
}
func main() {
	//fname := "./data/006331.csv"
	//fname := "./data/000307.csv"
	fname := "./data/005669.csv"
	//fname := "./data/efunds_sh50etf_110003.csv"
	//fname := "./data/000536.csv"
	//fname := "./data/531020.csv"

	//fh, error := os.Open("./data/efunds_sh50etf_110003.csv")

	fmt.Printf("filename: %s\n", fname)
	fh, error := os.Open(fname)
	//fh, error := os.Open("./data/531020.csv")
	//fh, error := os.Open("./data/005918.csv")
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
	//fmt.Println(fLine[0])

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

			//fmt.Printf(" the result of fundDataItem: %+v\n", fdi)
			fundData = append(fundData, fdi)

		}
	}

	// 买，卖加成比率

	var bonusRatioList = []BonousRatioPair{
		{decimal.NewFromFloat(0), decimal.NewFromFloat(1.0)},
		{decimal.NewFromFloat(1.0), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(1.0), decimal.NewFromFloat(1.0)},
		{decimal.NewFromFloat(0.0), decimal.NewFromFloat(0.0)},
		{decimal.NewFromFloat(50.0), decimal.NewFromFloat(50.0)},
		{decimal.NewFromFloat(50.0), decimal.NewFromFloat(0.382)},
		{decimal.NewFromFloat(50.0), decimal.NewFromFloat(0.1)},
		{decimal.NewFromFloat(0.999), decimal.NewFromFloat(0.125)},
		{decimal.NewFromFloat(37.5), decimal.NewFromFloat(37.5)},
		{decimal.NewFromFloat(0.618), decimal.NewFromFloat(0.382)},
		{decimal.NewFromFloat(0.125), decimal.NewFromFloat(50.0)},
		{decimal.NewFromFloat(150), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(100), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(50), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(25), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(20), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(10), decimal.NewFromFloat(0)},
		{decimal.NewFromFloat(5), decimal.NewFromFloat(0)},
	}

	for fi := 0.00; fi < 4.99; fi += 0.11 {
		for fj := 0.00; fj < 4.99; fj += 0.11 {

			bonusRatioList = append(bonusRatioList, BonousRatioPair{
				Buy:  decimal.NewFromFloat(fi),
				Sell: decimal.NewFromFloat(fj),
			})
		}
	}

	for _, b := range bonusRatioList {
		calcProfit(b, fundData)
	}

}

func calcProfit(b BonousRatioPair, fundData []FundDataItem) {
	//从2021年1月4日开始算

	var ab = &AccountBook{
		CostAccount:     0,
		EarnAccount:     0,
		FundAccount:     0,
		PlatformAccount: 0,
	}

	//买入基金操作

	//1. 从用户支出账户上扣除费用const
	//2. 计入手续费到 基金公司账户
	//3. 计入基金份额基金市值到  个人的基金账户
	//4. 计入入账到 基金公司的账户上

	//赎回基金操作
	//1. 从个人基金账户扣除自己
	//2. 计入手续费到到 基金公司账户
	//3. 记录支出 从 基金公司账户上
	//4. 记录入账到 个人的收益账户上

	//操作回溯，数据是从最新到最老，所以我们取数据，是从最底层开始取

	//操作算法： 如果当天对比前一天，是正的。那么取收益部分的 0.632； 如果是亏损，定额买入10元每天

	//这样操作一年下来，看最后我们是赚了，还是亏了。

	var accountLogList []AccountLog

	var todayProfit int64 = fundData[0].DayProfit

	for i := len(fundData) - 1; i >= 0; i-- {
		//		fmt.Printf("data: %+v", fundData[i])

		iData := fundData[i]
		if i == len(fundData)-1 {
			//		fmt.Println("第一次交易")

			nextTradeId := GetNextTradeId()

			realData := fundData[len(fundData)-2]
			//1. 从用户支出账户上扣除费用const
			ab.CostAccount = AsInt64(decimal.NewFromInt(0).Sub(startBalance))
			accountLogList = append(accountLogList, AccountLog{
				Id:          GetNextId(),
				TradeId:     nextTradeId,
				Date:        iData.Date,
				AccountType: 1,
				Amount:      -AsInt64(startBalance),
				AfterAmount: ab.CostAccount,
				Desc:        "扣除费用",
			})
			//2. 计入手续费到 基金公司账户
			rawFee := startBalance.Mul(feeRatio)
			fee := AsInt64(rawFee)
			ab.PlatformAccount = fee
			accountLogList = append(accountLogList, AccountLog{
				Id:          GetNextId(),
				TradeId:     nextTradeId,
				Date:        iData.Date,
				AccountType: 4,
				Amount:      fee,
				AfterAmount: ab.PlatformAccount,
				Desc:        "基金公司，手续费收入",
			})
			//3. 计入基金份额基金市值到  个人的基金账户
			rawChargeAmount := startBalance.Sub(rawFee)
			chargeAmount := AsInt64(rawChargeAmount)
			//除以今天的单价，等于基金份额
			ab.FundAccount = AsInt64(rawChargeAmount.Div(decimal.NewFromInt(realData.DayProfit)).Mul(ONEW))

			accountLogList = append(accountLogList, AccountLog{
				Id:             GetNextId(),
				TradeId:        nextTradeId,
				Date:           iData.Date,
				AccountType:    3,
				Amount:         chargeAmount,
				AfterAmount:    ab.FundAccount,
				Desc:           "个人基金账户，基金买入",
				ProfitExchange: float64(realData.DayProfit) / float64(10000),
			})
			//4. 计入入账到 基金公司的账户上
			ab.PlatformAccount += chargeAmount
			accountLogList = append(accountLogList, AccountLog{
				Id:          GetNextId(),
				TradeId:     nextTradeId,
				Date:        iData.Date,
				AccountType: 4,
				Amount:      chargeAmount,
				AfterAmount: ab.PlatformAccount,
				Desc:        "基金公司，用户买入基金",
			})

		} else {
			//上一次交易的值
			j := i + 1

			k := i - 1
			if k >= 0 {

				// realData 是次日的价格，一般基金购买，都会是次日结算，所以要用次日的价格来做账
				realData := fundData[k]
				oldData := fundData[j]

				if oldData.DayProfit == iData.DayProfit {
					// 不处理
				} else if oldData.DayProfit > iData.DayProfit {
					// 昨天的比今天的要高，今天亏钱了，补仓

					if b.Buy.Cmp(decimal.Zero) > 0 {
						diffCalc := decimal.NewFromInt(oldData.DayProfit).Sub(decimal.NewFromInt(iData.DayProfit))
						x := diffCalc.Div(decimal.NewFromInt(oldData.DayProfit))
						//		fmt.Printf(" 算出来的比率: %+v\n", x)
						// 当前账户
						//fmt.Printf("%+v\n", ab)
						ab, accountLogList = DoBuy(iData, realData, ab, accountLogList, x, b.Buy)
					}
				} else {
					//赚钱了，卖出

					if b.Sell.Cmp(decimal.Zero) > 0 {

						diffCalc := decimal.NewFromInt(iData.DayProfit).Sub(decimal.NewFromInt(oldData.DayProfit))
						y := diffCalc.Div(decimal.NewFromInt(oldData.DayProfit))
						//			fmt.Printf(" 算出来的比率: %+v\n", x)
						ab, accountLogList = DoSell(iData, realData, ab, accountLogList, y, b.Sell)
					}

				}
			}
		}

	}

	//WellPrint(accountLogList)

	x, _ := decimal.NewFromInt(ab.CostAccount).Div(ONEW).Float64()
	x5 := decimal.NewFromInt(ab.EarnAccount).Div(ONEW)
	x6 := decimal.NewFromInt(ab.FundAccount).Div(ONEW).Mul(decimal.NewFromInt(todayProfit)).Div(ONEW)
	rawX1 := x5.Add(x6)
	x1, _ := rawX1.Float64()

	rawX2 := rawX1.Sub(decimal.NewFromFloat(math.Abs(x)))
	x2, _ := rawX2.Float64()

	rawXr := rawX2.Div(decimal.NewFromFloat(math.Abs(x)))

	xr, _ := rawXr.Float64()

	const padding = 0
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
	x3, _ := b.Buy.Float64()
	x4, _ := b.Sell.Float64()
	fmt.Fprintf(w, "买入加成比率：%.4f\t卖出加成比率: %.4f\t账户：支出：%.4f\t收入合计：%.4f\t剪刀差: %.4f\t剪刀收益率: %.4f\t[收益: %.4f\t个人基金账户: %.4f份基金]\t基金公司账户：%.4f\t\n",
		x3,
		x4,

		x,
		x1,
		x2,
		xr,
		float64(ab.EarnAccount)/float64(10000),
		float64(ab.FundAccount)/float64(10000),
		float64(ab.PlatformAccount)/float64(10000))
	w.Flush()
}

func DoSell(f FundDataItem, g FundDataItem, ab *AccountBook, accountLogList []AccountLog, i decimal.Decimal, bonusRatio decimal.Decimal) (*AccountBook, []AccountLog) {
	//1. 从用户支出账户上扣除费用const

	rawSellAmount := decimal.NewFromInt(ab.FundAccount).Mul(i).Mul(bonusRatio)
	sellAmount := AsInt64(rawSellAmount)

	rawSellShare := rawSellAmount.Div(decimal.NewFromInt(g.DayProfit)).Mul(ONEW)
	sellShare := AsInt64(rawSellShare)

	if sellShare < 10000 {
		sellShare = 10000
	}

	if sellShare > ab.FundAccount {
		//fmt.Printf("sellShare: %f, FundAccount: %f卖出份额小于账户可卖份额，不处理\n", float64(sellShare)/float64(10000), float64(ab.FundAccount)/float64(10000))
		return ab, accountLogList
	}

	nextTradeId := GetNextTradeId()
	ab.FundAccount -= sellShare
	accountLogList = append(accountLogList, AccountLog{
		Id:             GetNextId(),
		TradeId:        nextTradeId,
		Date:           f.Date,
		AccountType:    3,
		Amount:         -sellShare,
		AfterAmount:    ab.FundAccount,
		Desc:           "基金卖出",
		ProfitExchange: float64(g.DayProfit) / float64(10000),
	})
	//2. 计入手续费到 基金公司账户
	rawFee := rawSellAmount.Mul(feeRatio)
	fee := AsInt64(rawFee)
	ab.PlatformAccount += fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 4,
		Amount:      fee,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，手续费收入",
	})
	//3. 计入基金份额基金市值到  个人的基金账户

	rawChargeAmount := rawSellAmount.Sub(rawFee)
	chargeAmount := AsInt64(rawChargeAmount)

	ab.EarnAccount += chargeAmount

	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 2,
		Amount:      chargeAmount,
		AfterAmount: ab.EarnAccount,
		Desc:        "个人收益账户，基金卖出入账",
	})
	//4. 计入入账到 基金公司的账户上
	ab.PlatformAccount -= sellAmount - fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 4,
		Amount:      chargeAmount,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，用户卖出基金",
	})

	return ab, accountLogList
}

func WellPrint(pl []AccountLog) {
	/**
	Id             int64
	Date           string
	AccountType    int
	Amount         int64
	AfterAmount    int64
	ProfitExchange float64
	Desc           string
	*/

	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "Id\tTradeId\tDate\tAccountType\tAmount\tAfterAmount\tProfitExchange\tDesc\t")
	for _, v := range pl {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t\n", v.Id, v.TradeId, v.Date, GetAccountType(v.AccountType), v.Amount, v.AfterAmount, v.ProfitExchange, v.Desc)
	}
	w.Flush()
}
func GetAccountType(n int) string {
	switch n {
	case 1:
		return "支出账户"
	case 2:
		return "收益账户"
	case 3:
		return "个人基金账户"
	case 4:
		return "基金公司账户"
	}
	return "无"
}
func DoBuy(f FundDataItem, g FundDataItem, ab *AccountBook, accountLogList []AccountLog, i decimal.Decimal, bonusRatio decimal.Decimal) (*AccountBook, []AccountLog) {
	//1. 从用户支出账户上扣除费用const

	rawBuyAmount := decimal.NewFromInt(ab.FundAccount).Mul(i).Mul(bonusRatio)
	buyAmount := AsInt64(rawBuyAmount)

	if buyAmount < 100000 {
		buyAmount = 100000
	}

	nextTradeId := GetNextTradeId()

	ab.CostAccount = ab.CostAccount - buyAmount
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 1,
		Amount:      -buyAmount,
		AfterAmount: ab.CostAccount,
		Desc:        "扣除费用",
	})
	//2. 计入手续费到 基金公司账户
	rawFee := rawBuyAmount.Mul(feeRatio)

	fee := AsInt64(rawFee)
	ab.PlatformAccount = ab.PlatformAccount + fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 4,
		Amount:      fee,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，手续费收入",
	})
	//3. 计入基金份额基金市值到  个人的基金账户
	rawChargeAmount := rawBuyAmount.Sub(rawFee)
	chargeAmount := AsInt64(rawChargeAmount)

	rawBuyShare := rawChargeAmount.Div(decimal.NewFromInt(g.DayProfit)).Mul(ONEW)

	buyShare := AsInt64(rawBuyShare)
	ab.FundAccount += buyShare

	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 3,
		Amount:      buyShare,
		AfterAmount: ab.FundAccount,
		Desc:        "个人基金账户，基金买入",
	})
	//4. 计入入账到 基金公司的账户上
	ab.PlatformAccount += chargeAmount
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		TradeId:     nextTradeId,
		Date:        f.Date,
		AccountType: 4,
		Amount:      chargeAmount,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，用户买入基金",
	})
	return ab, accountLogList
}
