package main

import (
	"bufio"
	"fmt"
	"github.com/grrrben/ip"
	"github.com/grrrben/latlong"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"os"
)

var rainCollection [24]Raintime
var printdata map[int][]string

type Raintime struct {
	Time      string
	Intensity uint8
}

type Raindata interface {
	mmpu() float32
}

func (r Raintime) mmpu() float64 {
	// Neerslagintensiteit = 10^((waarde-109)/32)
	// Ter controle: een waarde van 77 is gelijk aan een neerslagintensiteit van 0,1 mm/u.
	intensity := math.Pow(10, (float64(r.Intensity)-109)/32)
	return intensity
}

func main() {
	printdata = make(map[int][]string)
	ipaddress, err := ip.GetIp()
	if err != nil {
		fmt.Printf("Error: %q", err)
	}

	geo, err := latlong.Latlong(ipaddress)

	if err != nil {
		fmt.Print("-----------------------------------------\n")
		fmt.Println("Unable to get Geodata")
		fmt.Printf("%s\n", err)
		fmt.Print("-----------------------------------------\n")
		os.Exit(3)
	}

	weatherurl := fmt.Sprintf("https://br-gpsgadget-new.azurewebsites.net/data/raintext?lat=%s&lon=%s", geo.Latstring(), geo.Lngstring())

	resp, err := http.Get(weatherurl)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("Cannot get rain data")
	}

	parse(resp.Body)
	setPrintData()

	fmt.Print("-----------------------------------------\n")
	fmt.Print("Actual rain forecast in millimeters\n")
	if geo.City == "" {
		fmt.Printf("Precision: %s.\n", geo.GetCountry())
	} else {
		fmt.Printf("Precision: %s.\n", geo.City)
	}
	//fmt.Printf("%s\n", weatherurl)

	fmt.Print("-----------------------------------------\n")
	printonscreen()

	fmt.Print("-----------------------------------------\n")

}

func parse(body io.Reader) {
	it := 0
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		arr := strings.Split(scanner.Text(), "|")  // 000|10:15
		u, err := strconv.ParseUint(arr[0], 10, 8) // always gives an int64...
		if err != nil {
			fmt.Printf("could not parse %s", arr[0])
		}

		raintime := Raintime{Time: arr[1], Intensity: uint8(u)}

		rainCollection[it] = raintime
		it++
	}
}

func setPrintData() {

	printMinutes := [2]string{"00", "30"}

	it := 0

	for _, el := range rainCollection {
		temp := []string{}
		time := strings.Split(el.Time, ":")
		minutes := time[1]
		if it == 0 || stringInSlice(minutes, printMinutes) {
			for _, r := range el.Time {
				char := string(r)
				temp = append(temp, char)
			}
			printdata[it] = makeString10([]string{})
			it++
			printdata[it] = makeString10(temp)
			it++
			printdata[it] = makeString10([]string{})
			it++
			printdata[it] = makeString10(showRain(el))
			it++

		} else {
			printdata[it] = makeString10(showRain(el))
			it++
		}
	}
}

func showRain(el Raintime) []string {
	mmpu := []string{}
	mmu := round(el.mmpu(), 1, 0)
	for i := 0; i <= 10; i++ {
		if mmu > float64(i) {
			mmpu = append(mmpu, "#")
		} else if (mmu * 10) > float64(i) {
			mmpu = append(mmpu, "=")
		} else if (mmu * 100) > float64(i) {
			mmpu = append(mmpu, "-")
		}
	}
	return mmpu
}

func makeString10(str []string) []string {
	curLen := len(str)
	str10 := []string{}

	for i := 0; i <= (10 - curLen); i++ {
		str10 = append(str10, " ")
	}

	for i := 0; i <= len(str)-1; i++ {
		str10 = append(str10, str[i])
	}
	return str10
}

func printonscreen() {
	length := len(printdata)

	for i := 0; i <= 10; i++ {

		for j := 0; j <= (length - 1); j++ {
			fmt.Printf("%s", printdata[j][i])

		}
		fmt.Print("\n")
	}
}

func stringInSlice(a string, list [2]string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
