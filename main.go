package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "strings"
  "time"
  "sync"
)

func main() {
  //метка о начале работы, для того чтобы проверить скорость работы программы
  startTime := time.Now().UnixNano();
  //строка, котороую нужно найти
  neededString := "go"
  //максимальное количество потоков для http запросов
  streams := 5
  //канал для передачи данных в потоках
  mainChan := make(chan string)
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
  sliceLenght := len(sliceUrls)
  if (sliceLenght < streams) {
    streams = sliceLenght
  }


  //ограничитель потоков
  var wg sync.WaitGroup
  //устанавливаем количество потоков в ограничитель
  wg.Add(streams)

  //создаем отдельный поток, до максимального количества потоков
  for i := 1; i <= streams; i++ {
    go streamer(mainChan, &wg, neededString, &counterTotal)
  }

  // передаем каждый урл в канал
  for _, url := range sliceUrls {
      mainChan <- url
  }
  //закрываем канал
  close(mainChan)
  //ждем пока все каналы отработают
  wg.Wait()
  //выводим сумму совпадений в консоль
  fmt.Println("Total:", counterTotal)
  //метка времени окончания работы программы
  endTime := time.Now().UnixNano();
  //вывод в консоль времени работы программы
  fmt.Println(endTime - startTime)
}

//оснавная функция, которая запускается в потоки
func streamer(mainChan chan string, wg *sync.WaitGroup, neededString string, counterTotal *int) {
  //слушает все передачи url в канал
  for {
      url, more := <-mainChan
      //если это был последний переденный элемент, то перестаем ждать роботу потоков и выходим из функции
      if more == false {
          wg.Done()
          return
      }
      //получаем сумму совпадений
      count := sendAndCount(url, neededString)
      //добавляем сумму совпадений в счетчик
      *counterTotal += count
  }
}

//функкия объединяющая запрос по http и подсчет совпадений с выводом информации
func sendAndCount(url string, neededString string) int {
  //отправка http запроса и возврат тела ответа в виде строки
  body := curlSender(url)
  //вывод денных о количестве совпадений на url и возвращение количества совпадений
  return resultPrinter(neededString, body, url)
}

//отправка http запроса и возвращение ответа в виде строки
func curlSender(url string) string {
  //формирование GET зарпоса
  req, _ := http.NewRequest("GET", url, nil)
  //отрпавка сформированного GET запроса
  res, _ := http.DefaultClient.Do(req)
  //отложенное закрытие ресурса отправки, после выхода из функции
  defer res.Body.Close()
  //считываение тела ответа
  body, _ := ioutil.ReadAll(res.Body)
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
