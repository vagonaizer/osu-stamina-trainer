package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"osu-stamina-trainer/internal/menu"
	"strconv"
	"time"

	"github.com/mattn/go-tty"
)

func main() {

	menu.DisplayMenu()
	fmt.Println("CLI-Osu-Stamina-Trainer-v.0.0.1")
	fmt.Println("\n\n\n")

	keys := userSelectKeys()
	if len(keys) != 2 {
		log.Fatal("Необходимо выбрать ровно 2 клавиши.")
	}

	duration := userSelectDuration()

	fmt.Printf("Вы выбрали клавиши: %q и %q.\n", keys[0], keys[1])
	fmt.Printf("Программа будет считать нажатия в течение %v...\n", duration)
	fmt.Println("Нажимайте выбранные клавиши.")

	total, intervalsMs, totalMs, err := jsLikeClicker(keys, duration)
	if err != nil {
		log.Fatal(err)
	}

	// Конвертируем миллисекунды в секунды для Tap Speed
	seconds := float64(totalMs) / 1000.0

	// Tap Speed: "X taps / Y seconds"
	fmt.Printf("Tap Speed: %d taps / %.3f seconds.\n", total, seconds)

	// Stream Speed (BPM) по формуле из JS кода:
	// ((clickTimes.length) / (total_time_ms)) * 60000 / 4
	var streamBPM float64
	if totalMs > 0 {
		streamBPM = (float64(total) / float64(totalMs)) * 60000.0 / 4.0
	} else {
		streamBPM = 0.0
	}

	// Unstable Rate:
	// Из JS:
	// avg = среднее timediffs
	// std = sqrt(variance)
	// unstableRate = std * 10
	// Примечание: timediffs – интервалы между кликами в мс
	var unstableRateStr string
	if len(intervalsMs) < 2 {
		unstableRateStr = "n/a"
	} else {
		avg := mean(intervalsMs)
		var deviations []float64
		for _, v := range intervalsMs {
			diff := v - avg
			deviations = append(deviations, diff*diff)
		}
		variance := sum(deviations)
		variance = variance / float64(len(deviations))
		std := math.Sqrt(variance)
		unstableRate := std * 10.0
		unstableRateStr = fmt.Sprintf("%.3f", unstableRate)
	}

	fmt.Printf("Stream Speed: %.2f bpm\n", streamBPM)
	fmt.Printf("Unstable Rate: %s\n", unstableRateStr)
}

func jsLikeClicker(keys []rune, duration time.Duration) (int, []float64, int64, error) {
	t, err := tty.Open()
	if err != nil {
		return 0, nil, 0, err
	}
	defer t.Close()

	runeChan := make(chan rune)
	go func() {
		defer close(runeChan)
		for {
			r, err := t.ReadRune()
			if err != nil {
				return
			}
			runeChan <- r
		}
	}()

	startTime := time.Now().UnixNano() / 1_000_000 // ms
	count := 0
	var clickTimes []int64 // хранение времени нажатия в мс
	timer := time.After(duration)

loop:
	for {
		select {
		case r, ok := <-runeChan:
			if !ok {
				break loop
			}
			for _, k := range keys {
				if r == k {
					// Записываем текущее время клика в мс
					nowMs := time.Now().UnixNano() / 1_000_000
					count++
					clickTimes = append(clickTimes, nowMs)
					break
				}
			}
		case <-timer:
			break loop
		}
	}

	endTime := time.Now().UnixNano() / 1_000_000
	totalMs := endTime - startTime

	// Вычисляем интервалы между нажатиями в мс
	var intervalsMs []float64
	if len(clickTimes) > 1 {
		for i := 1; i < len(clickTimes); i++ {
			diff := float64(clickTimes[i] - clickTimes[i-1]) // уже в мс
			intervalsMs = append(intervalsMs, diff)
		}
	}

	return count, intervalsMs, totalMs, nil
}

// userSelectKeys запрашивает у пользователя две клавиши для отслеживания.
func userSelectKeys() []rune {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Введите первую клавишу: ")
	scanner.Scan()
	line1 := scanner.Text()

	fmt.Print("Введите вторую клавишу: ")
	scanner.Scan()
	line2 := scanner.Text()

	var keys []rune
	if len(line1) > 0 {
		keys = append(keys, []rune(line1)[0])
	}
	if len(line2) > 0 {
		keys = append(keys, []rune(line2)[0])
	}

	return keys
}

// userSelectDuration запрашивает у пользователя время в секундах.
func userSelectDuration() time.Duration {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Введите время в секундах: ")
	scanner.Scan()
	line := scanner.Text()

	secs, err := strconv.Atoi(line)
	if err != nil || secs <= 0 {
		log.Fatal("Неверный формат времени. Введите положительное целое число.")
	}

	return time.Duration(secs) * time.Second
}

func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	return sum(data) / float64(len(data))
}

func sum(data []float64) float64 {
	s := 0.0
	for _, v := range data {
		s += v
	}
	return s
}
