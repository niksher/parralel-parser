package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "strings"
  "time"
)

func main() {
  //метка о начале работы, для того чтобы проверить скорость работы программы
  startTime := time.Now().UnixNano();
  //строка, котороую нужно найти
  const NeededString = "go"
  //максимальное количество потоков для http запросов
  streams := 5
  //канал для передачи url в потоках
  mainChan := make(chan string)
  defer close(mainChan)
  //канал для передачи количества найденных совпадений в потоках
  counterChan := make(chan int)
  defer close(counterChan)
  //счетчик количества совпадений
  counterTotal := 0

  //слайс с url на которых нужно искать совпадения
  sliceUrls := []string {
    "https://golang.org",
    "https://go.dev/learn",
    "https://go.dev/doc",
    "https://go.dev/blog",
    "https://go.dev/blog/intro-generics",
    "https://go.dev/blog/go1.18beta2",
    "https://go.dev/blog/survey2022-q2"}

  //если количество урлов меньше максимального количества потоков программы, то ограничиваем максимальное количество потоков до количества url
  sliceLength := len(sliceUrls)
  if (sliceLength < streams) {
    streams = sliceLength
  }

  //создаем отдельный поток, до максимального количества потоков
  for i := 1; i <= streams; i++ {
    go streamer(mainChan, counterChan, NeededString)
  }

  // передаем каждый урл в канал
  for _, url := range sliceUrls {
    mainChan <- url
  }
  /*for {
    countString := <- counterChan
    counterTotal += countString
  }*/
  for countString := range counterChan {
    counterTotal += countString
  }
  //выводим сумму совпадений в консоль
  fmt.Println("Total:", counterTotal)
  //метка времени окончания работы программы
  endTime := time.Now().UnixNano();
  //вывод в консоль времени работы программы
  fmt.Println(endTime - startTime)
}

//основная функция, которая запускается в потоки
func streamer(mainChan chan string, counterChan chan int, neededString string) {
  //слушает все передачи url в канал
  for {
      url, more := <-mainChan
      //если это был последний переденный элемент, то перестаем ждать р работу потоков и выходим из функции
      if more == false {
        return
      }
      //получаем сумму совпадений
      count := sendAndCount(url, neededString)
      //добавляем сумму совпадений в счетчик
      counterChan <- count
  }
}

//функция объединяющая запрос по http и подсчет совпадений с выводом информации
func sendAndCount(url string, neededString string) int {
  //отправка http запроса и возврат тела ответа в виде строки
  body := curlSender(url)
  //вывод денных о количестве совпадений на url и возвращение количества совпадений
  return resultPrinter(neededString, body, url)
}

//отправка http запроса и возвращение ответа в виде строки
func curlSender(url string) string {
  //формирование GET зарпоса
  req, reqErr := http.NewRequest("GET", url, nil)
  if (reqErr != nil) {
    fmt.Printf("Request error: %v\n", reqErr)
  }
  //отрпавка сформированного GET запроса
  res, resErr := http.DefaultClient.Do(req)
  if (resErr != nil) {
    fmt.Printf("Response error: %v\n", resErr)
  }
  //отложенное закрытие ресурса отправки, после выхода из функции
  defer res.Body.Close()
  //считываение тела ответа
  body, readErr := ioutil.ReadAll(res.Body)
  if (readErr != nil) {
    fmt.Printf("Read response body error: %v\n", readErr)
  }
  //возвращение тела ответа в виде строки
  return string(body)
}

//счетчик совпадений искомой строки в теле ответа и вывод результата в консоль
func resultPrinter(neededString string, body string, url string) int {
  //счетчик совпадений искомой строки в теле ответа
  countString := strings.Count(body, neededString)
  //вывод результата в консоль
  fmt.Println("Count string \"" + neededString + "\" in url", url, "is equal:", countString)
  //возвращение количества совпадений
  return countString
}
