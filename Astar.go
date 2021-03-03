package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)


type ViewData struct{
	Columns int
	Rows int

}

type Path struct{
	Columns int
	Rows int
}


type Star struct{
	i int
	j int
	f int
	g int
	h int
	neighbors []*Star
	previous *Star
	wall bool

}
var colomns int = 50
var rows int = 50


//var path []Star


var mux = http.NewServeMux()
var tpl = template.Must(template.ParseFiles("index.html"))


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}


func contains(s []Star, e Star) bool {
	for _, a := range s {
		if ((a.i == e.i) && (a.j == e.j)) {
			return true

		}
	}

	return false
}


func find(s []Star, e Star) int {
	for i := 0;i < len(s);i++ {
		if ((s[i].i == e.i) && (s[i].j == e.j)) {
			return i
		}
	}
	return 0
}



func heuristic(a Star,b Star) float64{

	var d = math.Round( math.Sqrt( math.Pow(math.Abs(float64(a.i-b.i)),2) +  math.Pow(math.Abs(float64(a.j-b.j)),2)  ))
	//var d = math.Abs(float64(a.i-b.i)) + math.Abs(float64(a.j-b.j))
	return d

}

var walls []Star

func reader(conn *websocket.Conn){
	defer conn.Close()
	for{
		var tmp Star
		_,p,err := conn.ReadMessage()
		if(err!=nil){
			fmt.Println(err, "jopa")
			return
		} else{
			if(string(p)[0] == 99) {
				str := strings.Split(string(p), "c")
				str1 := strings.Split(str[1], " ")
				i, _ := strconv.Atoi(str1[0])
				j, _ := strconv.Atoi(str1[1])
				tmp.i = i
				tmp.j = j
				walls = append(walls, tmp)
			}
			if(string(p)[0] == 100) {
				str := strings.Split(string(p), "d")
				str1 := strings.Split(str[1], " ")
				i, _ := strconv.Atoi(str1[0])
				j, _ := strconv.Atoi(str1[1])
				tmp.i = i
				tmp.j = j
				var del = find(walls, tmp)
				walls = append(walls[:del], walls[del+1:]...)

			}
		}




	}

}



var randomWals = false

var count = 0



func Astar(w http.ResponseWriter, r *http.Request){

	var Result []Star

	var openSet  []Star
	var closedSet []Star
	var start Star
	var end Star



	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil{
		log.Print(err)
	}



	//data := ViewData{
	//	Columns: colomns,
	//	Rows: rows,
	//}

	grid := make([][]Star, colomns)
	for i := range grid {
		grid[i] = make([]Star, rows)
	}



	for i := 0; i < colomns;i++{
		for j := 0; j < rows;j++{
			grid[i][j].i = i
			grid[i][j].j = j
			grid[i][j].f = 0
			grid[i][j].g = 0
			grid[i][j].h = 0
			grid[i][j].previous = new(Star)
			grid[i][j].previous.i = -1
			grid[i][j].wall = false

			if (randomWals == true){

				if(rand.Int31n(10) < 4) {
					grid[i][j].wall = true
					walls = append(walls, grid[i][j])
					count = 3

				}
			}

		}
	}


	randomWals = false


	go reader(ws)



	fmt.Println(len(walls))
	for i := 0; i < len(walls) ;i++ {
		//grid[walls[i].i][walls[i].j].wall = true
		str1 := "w" + strconv.Itoa(walls[i].i) + " " + strconv.Itoa(walls[i].j)
		err = ws.WriteMessage(1, []byte(str1))
		if err != nil {
			fmt.Print(err)
		}
	}


	str := "col" + strconv.Itoa(colomns) + " " + strconv.Itoa(rows)
	err = ws.WriteMessage(1, []byte(str))
	if err != nil{
		fmt.Print(err)
	}


	for i := 0; i < colomns;i++ {
		for j := 0; j < rows; j++ {
			//grid[i][j].neighbors = new()
			if(i < colomns-1){

				grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i+1][j])
			}
			if (i>0){
				grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i-1][j])
			}
			if(j < rows-1){
				grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i][j+1])
			}
			if(j > 0 ){
				grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i][j-1])
			}


			if(i < colomns-1 && j < rows-1 ){
				if (contains(walls,grid[i+1][j]) && contains(walls,grid[i][j+1])){

				}else {
					grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i+1][j+1])
				}
			}



			if(i > 0 && j > 0){
				if (contains(walls,grid[i-1][j]) && contains(walls,grid[i][j-1])){

				}else {
					grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i-1][j-1])
				}

			}

			if(i < colomns-1 && j >0){
				if (contains(walls,grid[i+1][j]) && contains(walls,grid[i][j-1])){

				}else {
					grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i+1][j-1])
				}


			}

			if(i > 0 && j < rows-1){
				if (contains(walls,grid[i-1][j]) && contains(walls,grid[i][j+1])){

				}else {
					grid[i][j].neighbors = append(grid[i][j].neighbors, &grid[i-1][j+1])
				}


			}

		}
	}


	start = grid[0][0] //задаем точку входа
	start.g = 0
	start.f = int(heuristic(start,end))
	end = grid[colomns-1][rows-1] // задаем нужноую точку
	//fmt.Println(end.i, end.j)
	openSet = append(openSet, start)


	for len(openSet) != 0{
		//делаем

		//fmt.Println("jopa")
		var winner int = 0
		for i := 0; i < len(openSet);i++{
			//fmt.Print(openSet[i].f)
			if openSet[i].f < openSet[winner].f {
				winner = i
			}
		}
		var current = openSet[winner]
		//fmt.Println(current.g, winner)
		if (current.i == end.i) && (current.j == end.j) {
			var temp = current
			//fmt.Println(&current.previous.i, "dasdas")
			Result = append(Result, temp)
			for(temp.previous.i != -1){
				Result = append(Result, *temp.previous)
				temp = *temp.previous
			}
			for i := len(Result)-1; i >= 0;i-- {
				str1 := "p" + strconv.Itoa(Result[i].i) + " " + strconv.Itoa(Result[i].j)
				err = ws.WriteMessage(1, []byte(str1))
				if err != nil{
					fmt.Print(err)
				}
			}
			fmt.Print("DONE")
			break
		}
		//fmt.Println(openSet[winner].i, openSet[winner].j,len(openSet[winner].neighbors))
		var del = find(openSet, current)
		openSet = append(openSet[:del], openSet[del+1:]...)
		//fmt.Println(openSet)
		closedSet = append(closedSet, current)


		var neighbors = current.neighbors


		for  i:=0;i < len(neighbors);i++{

			var neighbor = neighbors[i]
			//fmt.Println(neighbor.i, neighbor.j)
			if (!contains(closedSet,*neighbor) && !neighbor.wall && !contains(walls, *neighbor)) {
				var tempG = current.g + 1

				if contains(openSet,*neighbor) {
					if tempG < neighbor.g {
						neighbor.g = tempG
					}
				} else {
					//fmt.Println(grid[0][1].neighbors[0].i, grid[0][1].neighbors[0].j)
					neighbor.g = tempG
					//fmt.Println(neighbor.f)
					neighbor.h = int(heuristic(*neighbor, end))
					neighbor.f = neighbor.g + neighbor.h
					*neighbor.previous = current
					openSet = append(openSet, *neighbor)

				}


			}
		}



		//
		//str1 := "j"
		//err = ws.WriteMessage(1, []byte(str1))
		//if err != nil{
		//	fmt.Print(err)
		//}
		//
		//time.Sleep(1 * time.Second)
		time.Sleep(50 * time.Millisecond)
		//path = nil

		for i := 0; i < len(openSet) ;i++ {
			str1 := "o" + strconv.Itoa(openSet[i].i) + " " + strconv.Itoa(openSet[i].j)
			err = ws.WriteMessage(1, []byte(str1))
			if err != nil {
				fmt.Print(err)
			}
		}
		for i := 0; i < len(closedSet) ;i++ {
			str1 := "e" + strconv.Itoa(closedSet[i].i) + " " + strconv.Itoa(closedSet[i].j)
			err = ws.WriteMessage(1, []byte(str1))
			if err != nil {
				fmt.Print(err)
			}
		}




		var path  []Star
		var temp = current
		//fmt.Println(&current.previous.i, "dasdas")
		path = append(path, temp)
		for(temp.previous.i != -1){
			path = append(path, *temp.previous)
			temp = *temp.previous
		}
		for i := len(path) -1; i >= 0;i-- {
			str1 := "p" + strconv.Itoa(path[i].i) + " " + strconv.Itoa(path[i].j)
			err = ws.WriteMessage(1, []byte(str1))
			if err != nil{
				fmt.Print(err)
			}
		}

		str1 := "j"
		err = ws.WriteMessage(1, []byte(str1))
	}
	if len(openSet) == 0{
		fmt.Println("no solution")
	}

	for i :=len(Result) - 1; i >= 0;i-- {
		fmt.Println(Result[i].i, Result[i].j)
	}



	//tpl.Execute(w, data)




	//defer ws.Close()

}

func setup(){
	http.HandleFunc("/astar", Astar)
}

func main(){
	setup()
	log.Println("Запуск веб-сервера на http://127.0.0.1:4000")
	err := http.ListenAndServe(":4000", nil)
	log.Fatal(err)
}