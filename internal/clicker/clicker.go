package clicker

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mattn/go-tty"
)

func Сlicker(keys []rune, duration time.Duration) (int, error) {
	t, err := tty.Open()
	if err != nil {
		return 0, err
	}
	defer t.Close()

	// Канал для передачи нажатых клавиш
	runeChan := make(chan rune)

	// Запускаем горутину для чтения клавиш
	go func() {
		defer close(runeChan)
		for {
			r, err := t.ReadRune()
			if err != nil {
				// Если ошибка чтения, завершаем горутину
				return
			}
			runeChan <- r
		}
	}()

	count := 0
	// Канал, который "сработает" по истечении времени
	timer := time.After(duration)

	for {
		select {
		case r, ok := <-runeChan:
			if !ok {
				// Канал закрылся, завершаем
				return count, nil
			}
			// Считаем только те клавиши, которые соответствуют выбранным
			for _, k := range keys {
				if r == k {
					count++
					break
				}
			}
		case <-timer:
			// Время истекло, завершаем подсчёт
			return count, nil
		}
	}
}

func UserSelectKeys() []rune {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Введите первую клавишу: ")
	scanner.Scan()
	line1 := scanner.Text()

	fmt.Print("Введите вторую клавишу: ")
	scanner.Scan()
	line2 := scanner.Text()

	// Предполагаем, что пользователь вводит одну клавишу (один символ)
	var keys []rune
	if len(line1) > 0 {
		keys = append(keys, []rune(line1)[0])
	}
	if len(line2) > 0 {
		keys = append(keys, []rune(line2)[0])
	}

	return keys
}

func UserSelectDuration() time.Duration {
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
