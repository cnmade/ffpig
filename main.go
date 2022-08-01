package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

type AccountLog struct {
	Id             int64
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
	Buy  float64
	Sell float64
}

// 手续费是0.01
var feeRatio = 0.0005

//一万块钱开始
var startBalance int64 = 10000 * 10000

//ID发号器
var IdChain int64 = 1

func GetNextId() int64 {
	nextId := atomic.LoadInt64(&IdChain) + 1
	atomic.StoreInt64(&IdChain, nextId)
	return nextId
}
func main() {

	fh, error := os.Open("./data/efunds_sh50etf_110003.csv")

	//fh, error := os.Open("./data/006331.csv")
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

	//var buyBonousRatio float64 = 0.999
	//var sellBonousRatio float64 = 0.125
	//var buyBonousRatio float64 = 37.5
	//var sellBonousRatio float64 = 37.5

	//var buyBonousRatio float64 = 0.618
	//var sellBonousRatio float64 = 0.382

	//var buyBonousRatio float64 = 0.125
	//var sellBonousRatio float64 = 50.0
	var bonusRatioList = []BonousRatioPair{
		{0.125, 50.0},
	}
	//var bonusRatioList = []BonousRatioPair{
	//	{0, 1.0},
	//	{1.0, 0},
	//	{1.0, 1.0},
	//	{0.0, 0.0},
	//	{50.0, 50.0},
	//	{50.0, 0.382},
	//	{50.0, 0.1},
	//	{0.999, 0.125},
	//	{37.5, 37.5},
	//	{0.618, 0.382},
	//	{0.125, 50.0},
	//	{150, 0},
	//	{100, 0},
	//	{50, 0},
	//	{25, 0},
	//	{20, 0},
	//	{10, 0},
	//	{5, 0},
	//}

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

			realData := fundData[len(fundData)-2]
			//1. 从用户支出账户上扣除费用const
			ab.CostAccount = -startBalance
			accountLogList = append(accountLogList, AccountLog{
				Id:          GetNextId(),
				Date:        iData.Date,
				AccountType: 1,
				Amount:      -startBalance,
				AfterAmount: ab.CostAccount,
				Desc:        "扣除费用",
			})
			//2. 计入手续费到 基金公司账户
			fee := int64(float64(startBalance) * feeRatio)
			ab.PlatformAccount = fee
			accountLogList = append(accountLogList, AccountLog{
				Id:          GetNextId(),
				Date:        iData.Date,
				AccountType: 4,
				Amount:      fee,
				AfterAmount: ab.PlatformAccount,
				Desc:        "基金公司，手续费收入",
			})
			//3. 计入基金份额基金市值到  个人的基金账户
			chargeAmount := startBalance - fee
			//除以今天的单价，等于基金份额
			ab.FundAccount = int64((float64(chargeAmount) / float64(realData.DayProfit)) * 10000)

			accountLogList = append(accountLogList, AccountLog{
				Id:             GetNextId(),
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

					if b.Buy > 0 {
						x := float64(oldData.DayProfit-iData.DayProfit) / float64(oldData.DayProfit)
						//		fmt.Printf(" 算出来的比率: %+v\n", x)
						// 当前账户
						fmt.Printf("%+v\n", ab)
						ab, accountLogList = DoBuy(iData, realData, ab, accountLogList, x, b.Buy)
					}
				} else {
					//赚钱了，卖出

					if b.Sell > 0 {
						x := float64(iData.DayProfit-oldData.DayProfit) / float64(oldData.DayProfit)

						//			fmt.Printf(" 算出来的比率: %+v\n", x)
						ab, accountLogList = DoSell(iData, realData, ab, accountLogList, x, b.Sell)
					}

				}
			}
		}

	}

	WellPrint(accountLogList)

	x := float64(ab.CostAccount) / float64(10000)
	x1 := float64(ab.EarnAccount)/float64(10000) + float64(ab.FundAccount)/float64(10000)*float64(todayProfit)/float64(10000)
	x2 := x1 - math.Abs(x)
	xr := float64(x2) / float64(math.Abs(x))
	fmt.Printf("买入加成比率：%f\t卖出加成比率: %f\t账户：支出：%f\t收入合计：%f\t剪刀差: %f\t剪刀收益率: %f\t[收益: %f\t个人基金账户: %f份基金]\t基金公司账户：%f\n",
		b.Buy,
		b.Sell,

		x,
		x1,
		x2,
		xr,
		float64(ab.EarnAccount)/float64(10000),
		float64(ab.FundAccount)/float64(10000),
		float64(ab.PlatformAccount)/float64(10000))
}

func DoSell(f FundDataItem, g FundDataItem, ab *AccountBook, accountLogList []AccountLog, i float64, bonusRatio float64) (*AccountBook, []AccountLog) {
	//1. 从用户支出账户上扣除费用const

	sellAmount := int64(float64(ab.FundAccount) * i * bonusRatio)
	sellShare := int64(float64(sellAmount) / float64(g.DayProfit) * 10000)

	ab.FundAccount -= sellShare
	accountLogList = append(accountLogList, AccountLog{
		Id:             GetNextId(),
		Date:           f.Date,
		AccountType:    3,
		Amount:         -sellShare,
		AfterAmount:    ab.FundAccount,
		Desc:           "基金卖出",
		ProfitExchange: float64(g.DayProfit) / float64(10000),
	})
	//2. 计入手续费到 基金公司账户
	fee := int64(float64(sellAmount) * feeRatio)
	ab.PlatformAccount += fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        f.Date,
		AccountType: 4,
		Amount:      fee,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，手续费收入",
	})
	//3. 计入基金份额基金市值到  个人的基金账户
	chargeAmount := sellAmount - fee

	ab.EarnAccount += chargeAmount

	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
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
	fmt.Println("Id\tDate\tAccountType\tAmount\tAfterAmount\tProfitExchange\tDesc")
	for _, v := range pl {
		fmt.Printf(" %v\t%v\t%v\t%v %v\t%v\t%v\n", v.Id, v.Date, GetAccountType(v.AccountType), v.Amount, v.AfterAmount, v.ProfitExchange, v.Desc)
	}
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
func DoBuy(f FundDataItem, g FundDataItem, ab *AccountBook, accountLogList []AccountLog, i float64, bonusRatio float64) (*AccountBook, []AccountLog) {
	//1. 从用户支出账户上扣除费用const

	buyAmount := int64(float64(ab.FundAccount) * i * bonusRatio)

	ab.CostAccount = ab.CostAccount - buyAmount
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        f.Date,
		AccountType: 1,
		Amount:      -buyAmount,
		AfterAmount: ab.CostAccount,
		Desc:        "扣除费用",
	})
	//2. 计入手续费到 基金公司账户
	fee := int64(float64(buyAmount) * feeRatio)
	ab.PlatformAccount = ab.PlatformAccount + fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        f.Date,
		AccountType: 4,
		Amount:      fee,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，手续费收入",
	})
	//3. 计入基金份额基金市值到  个人的基金账户
	chargeAmount := buyAmount - fee

	buyShare := int64(float64(chargeAmount) / float64(g.DayProfit) * 10000)
	ab.FundAccount += buyShare

	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
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
		Date:        f.Date,
		AccountType: 4,
		Amount:      chargeAmount,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，用户买入基金",
	})
	return ab, accountLogList
}
