```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

//Version directa
// func fetchURL(url string) {
// 	start := time.Now()
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		fmt.Printf("hay un error en el fetching %s: %v\n", url, err)
// 		return
// 	}
// 	defer resp.Body.Close()
// 	duracion := time.Since(start)
// 	fmt.Printf("Fetcheado %s en %v - Status %s \n", url, duracion, resp.Status)
// }

// Version Go routine
func fetchURL(url string, wg *sync.WaitGroup) {
	defer wg.Done() //marca go routine como completada.

	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("hay un error en el fetching %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	duracion := time.Since(start)
	fmt.Printf("Fetcheado %s en %v - Status %s \n", url, duracion, resp.Status)
}
func main() {

	urls := []string{
		"https://rickandmortyapi.com/api/character/1",
		"https://rickandmortyapi.com/api/character/2",
		"https://rickandmortyapi.com/api/character/3",
		"https://rickandmortyapi.com/api/character/4",
	}
	//Version directa
	// start := time.Now()
	// for _, url := range urls {
	// 	fetchURL(url)
	// }
	// duracionTotal := time.Since(start)
	// fmt.Printf("duracion total de la secuencia: %v \n", duracionTotal)
	// superavit := "https://superavitformacion.com/"
	// superavitAlumnos := "https://curso.superavitformacion.com/"
	// superavitEntrar := "https://curso.superavitformacion.com/courses/curso-superavit-tsd/"
	// fetchURL(superavit)
	// fetchURL(superavitAlumnos)
	// fetchURL(superavitEntrar)

	// Version con Go rutines
	start := time.Now()
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go fetchURL(url, &wg)
	}
	wg.Wait()
	duracionTotal := time.Since(start)
	fmt.Printf("Tiempo total concurrente de la secuencia: %v \n", duracionTotal)
}
```
