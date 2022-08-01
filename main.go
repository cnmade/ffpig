package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

type AccountLog struct {
	Id          int64
	Date        string
	AccountType int
	Amount      int64
	AfterAmount int64
	Desc        string
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

// 手续费是0.01
var feeRatio = 0.0015

//ID发号器
var IdChain int64 = 1

func GetNextId() int64 {
	nextId := atomic.LoadInt64(&IdChain) + 1
	atomic.StoreInt64(&IdChain, nextId)
	return nextId
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

	//一万块钱开始
	var startBalance int64 = 10000 * 10000

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

	for i := len(fundData) - 1; i >= 0; i-- {
		fmt.Printf("data: %+v", fundData[i])

		iData := fundData[i]
		if i == len(fundData)-1 {
			fmt.Println("第一次交易")

			//1. 从用户支出账户上扣除费用const
			ab.CostAccount = ab.CostAccount - startBalance
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
			ab.PlatformAccount = ab.PlatformAccount + fee
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
			ab.FundAccount = chargeAmount

			accountLogList = append(accountLogList, AccountLog{
				Id:          GetNextId(),
				Date:        iData.Date,
				AccountType: 3,
				Amount:      chargeAmount,
				AfterAmount: ab.FundAccount,
				Desc:        "个人基金账户，基金买入",
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
			oldData := fundData[j]

			if oldData.DayProfit == iData.DayProfit {
				// 不处理
			} else if oldData.DayProfit > iData.DayProfit {
				// 昨天的比今天的要高，今天亏钱了，补仓

				x := float64(oldData.DayProfit-iData.DayProfit) / float64(oldData.DayProfit)
				fmt.Printf(" 算出来的比率: %+v\n", x)
				ab, accountLogList = DoBuy(iData.Date, ab, accountLogList, x)
			} else {
				//赚钱了，卖出

				x := float64(iData.DayProfit-oldData.DayProfit) / float64(oldData.DayProfit)

				fmt.Printf(" 算出来的比率: %+v\n", x)
				ab, accountLogList = DoSell(iData.Date, ab, accountLogList, x)
			}
		}

	}

	fmt.Printf("账户日志:  %v\n", accountLogList)

	fmt.Printf("账户：支出：%f, 收入合计：%f,  [收益: %f, 个人基金账户: %f], 基金公司账户：%f\n", float64(ab.CostAccount)/float64(10000),
		float64(ab.EarnAccount)/float64(10000)+float64(ab.FundAccount)/float64(10000),
		float64(ab.EarnAccount)/float64(10000),
		float64(ab.FundAccount)/float64(10000),
		float64(ab.PlatformAccount)/float64(10000))

}

func DoSell(date string, ab *AccountBook, accountLogList []AccountLog, i float64) (*AccountBook, []AccountLog) {
	//1. 从用户支出账户上扣除费用const

	sellAmount := int64(float64(ab.FundAccount) * i * 0.9)

	ab.FundAccount -= sellAmount
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        date,
		AccountType: 3,
		Amount:      -sellAmount,
		AfterAmount: ab.FundAccount,
		Desc:        "基金卖出",
	})
	//2. 计入手续费到 基金公司账户
	fee := int64(float64(sellAmount) * feeRatio)
	ab.PlatformAccount += fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        date,
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
		Date:        date,
		AccountType: 2,
		Amount:      chargeAmount,
		AfterAmount: ab.EarnAccount,
		Desc:        "个人收益账户，基金卖出入账",
	})
	//4. 计入入账到 基金公司的账户上
	ab.PlatformAccount -= sellAmount - fee
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        date,
		AccountType: 4,
		Amount:      chargeAmount,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，用户卖出基金",
	})

	return ab, accountLogList
}

func DoBuy(date string, ab *AccountBook, accountLogList []AccountLog, i float64) (*AccountBook, []AccountLog) {
	//1. 从用户支出账户上扣除费用const

	buyAmount := int64(float64(ab.FundAccount) * i * 0.1)

	ab.CostAccount = ab.CostAccount - buyAmount
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        date,
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
		Date:        date,
		AccountType: 4,
		Amount:      fee,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，手续费收入",
	})
	//3. 计入基金份额基金市值到  个人的基金账户
	chargeAmount := buyAmount - fee
	ab.FundAccount += chargeAmount

	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        date,
		AccountType: 3,
		Amount:      chargeAmount,
		AfterAmount: ab.FundAccount,
		Desc:        "个人基金账户，基金买入",
	})
	//4. 计入入账到 基金公司的账户上
	ab.PlatformAccount += chargeAmount
	accountLogList = append(accountLogList, AccountLog{
		Id:          GetNextId(),
		Date:        date,
		AccountType: 4,
		Amount:      chargeAmount,
		AfterAmount: ab.PlatformAccount,
		Desc:        "基金公司，用户买入基金",
	})
	return ab, accountLogList
}
