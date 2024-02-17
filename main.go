package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Expression struct {
	Value  string
	Number int
	Status int
	Start  time.Time
	Finish time.Time
	Result int
	mu     sync.Mutex
}
type Data struct {
	ListExpr []*Expression
}

var data = Data{}
var c = make(chan *Expression)
var op = map[string]time.Duration{
	"+": time.Duration(1 * time.Second),
	"-": time.Duration(2 * time.Second),
	"*": time.Duration(3 * time.Second),
	"/": time.Duration(4 * time.Second),
}

func Check(s string) int {
	// проверка валидности выражения
	// возвращает статус
	re := regexp.MustCompile("^*[0-9-]+[0-9 *+-/]+[0-9]$")
	if re.MatchString(s) {
		return 200
	}
	return 400
}

func Parse(s string) []string {
	res := strings.Split(s, " ")
	return res

}
func Add(a, b string) int {
	x, _ := strconv.Atoi(a)
	y, _ := strconv.Atoi(b)
	time.Sleep(op["+"])
	return x + y
}
func Sub(a, b string) int {
	x, _ := strconv.Atoi(a)
	y, _ := strconv.Atoi(b)
	time.Sleep(op["-"])
	return x - y
}
func Mult(a, b string) int {
	x, _ := strconv.Atoi(a)
	y, _ := strconv.Atoi(b)
	time.Sleep(op["*"])
	return x * y
}
func Divis(a, b string) int {
	x, _ := strconv.Atoi(a)
	y, _ := strconv.Atoi(b)
	time.Sleep(op["/"])
	return int(x / y)
}
func Calculator(val *Expression) {
	val.mu.Lock()
	defer val.mu.Unlock()
	s := val.Value
	if val.Status == 200 {
		p := Parse(s)
		if p[1] == "+" {
			val.Result = Add(p[0], p[2])
			val.Finish = time.Now()
		}
		if p[1] == "-" {
			val.Result = Sub(p[0], p[2])
			val.Finish = time.Now()
		}
		if p[1] == "*" {
			val.Result = Mult(p[0], p[2])
			val.Finish = time.Now()
		}
		if p[1] == "/" {
			if p[2] == "0" {
				val.Status = 500
				val.Finish = time.Now()
			} else {
				val.Result = Divis(p[0], p[2])
				val.Finish = time.Now()
			}
		}
	} else {
		val.Finish = time.Now()
	}
}
func Arithmetic(w http.ResponseWriter, r *http.Request) {
	expr := r.FormValue("expression")
	status := Check(expr)
	n := 0
	flag := true
	for _, el := range data.ListExpr {
		if el.Value == expr {
			n = el.Number
			flag = false
			break
		}
	}
	if n == 0 && status == 200 {
		n = rand.Intn(1000)
	}
	m := Expression{
		Value:  expr,
		Number: n,
		Start:  time.Now(),
		Status: status,
	}
	if flag {
		data.ListExpr = append(data.ListExpr, &m)
	}

	c <- &m

	if status == 200 {
		fmt.Fprintf(w, "Ваше выражение %s имеет идентификатор %d и обрабатывается", expr, n)

	} else {
		fmt.Fprintf(w, "Ваше выражение %s невалидно, его статус %d", expr, status)
	}
}

func DataBase(w http.ResponseWriter, r *http.Request) {
	for _, el := range data.ListExpr {
		fmt.Fprintf(w, "Expression: %s,\tID: %d,\tStatus: %d,\tStartDate: %s,\tFinishDate: %s,\tResult: %d\n",
			el.Value, el.Number, el.Status, el.Start.Format("2006-01-02 15:04:05"), el.Finish.Format("2006-01-02 15:04:05"), el.Result)
	}
}

func Operations(w http.ResponseWriter, r *http.Request) {
	for key, val := range op {
		fmt.Fprintf(w, "Operation: %s\tDuration: %s\n", key, val)
	}
}

func main() {
	// обработчик формы ввода запроса
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "exprForm.html")
	})
	// обработчик показа ответа после ввода выражения с ID
	http.HandleFunc("/expr", http.HandlerFunc(Arithmetic))
	// обработчик показа списка выражений в базе выражений
	http.HandleFunc("/base", DataBase)
	// обработчик показа списка операций и времени выполнения
	http.HandleFunc("/oper", Operations)

	// функция вычислителя
	go func() {
		for r := range c {
			Calculator(r)
		}
	}()
	// обработчик показа списка операций
	http.ListenAndServe(":8000", nil)
}
