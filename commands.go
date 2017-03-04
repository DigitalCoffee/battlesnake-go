package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CellType uint8

const (
	EMPTY CellType = iota
	SNAKE
	FOOD
)

type Cell struct {
	t     CellType
	snake string
	pos   int
}

type Dir int8

const (
	UP Dir = iota
	RIGHT
	DOWN
	LEFT
	num_dirs
)

var directions = [4]string{
	UP:    "up",
	DOWN:  "down",
	LEFT:  "left",
	RIGHT: "right",
}

type TurnData struct {
	req     *MoveRequest
	board   [][]Cell
	mysnake *Snake
}

func abs(i int) int {
	if i < 0 {
		return -i
	} else {
		return i
	}
}

func heuristic_cost(start Point, end Point) int {
	return abs(start.X-end.X) + abs(start.Y-end.Y)

}

/*
func AStar(head Point, go_to Point, data TurnData) Point {

	// Reference: https://en.wikipedia.org/wiki/A*_search_algorithm#Pseudocode

	//Setup

	width := data.req.Width
	height := data.req.Height

	closedSet := []Point{}
	openSet := []Point{head}
	// cost to get to individual point
	// access point as go_score[i][j] = go_score[i*width + j]
	go_score := make([]int, height*width)
	go_score[start.y*width+start.x] = 0
	//cost of travelling to go_to through the point
	through_score := make([]int, height*width)
	through_score[start.y*width+start.x] = heuristic_cost(head, go_to)

	// Computation
	for len(openSet) > 0 {
		// current := the node in openSet having the lowest fScore[] value
	}
	return Point{-1, -1}
}*/

func reconstruct_path(from Point, current Point) Point {
	return Point{-1, -1}
}

func buildBoard(req *MoveRequest) (board [][]Cell) {
	board = make([][]Cell, req.Width)
	for i := range board {
		board[i] = make([]Cell, req.Height)
	}

	for _, food := range req.Food {
		board[food.Y][food.X].t = FOOD
	}

	for _, snake := range req.Snakes {
		for i, body := range snake.Coords {
			board[body.Y][body.X] = Cell{t: SNAKE, snake: snake.Id, pos: i}
		}
	}
	return board
}

func getSnake(req *MoveRequest, id string) Snake {
	for _, snake := range req.Snakes {
		if snake.Id == id {
			return snake
		}
	}
	return Snake{}
}

func safeMove(data *TurnData, dir Dir) int {
	req := data.req
	board := data.board
	myhead := data.mysnake.Coords[0]
	mylen := len(data.mysnake.Coords)

	testCells := make([]*Cell, 0, 3)

	if dir == UP && myhead.Y > 0 && board[myhead.Y-1][myhead.X].t != SNAKE {
		if myhead.X > 0 {
			testCells = append(testCells, &board[myhead.Y-1][myhead.X-1])
		}
		if myhead.Y > 1 {
			testCells = append(testCells, &board[myhead.Y-2][myhead.X])
		}
		if myhead.X < req.Width-1 {
			testCells = append(testCells, &board[myhead.Y-1][myhead.X+1])
		}
	} else if dir == DOWN && myhead.Y < req.Height-1 && board[myhead.Y+1][myhead.X].t != SNAKE {
		if myhead.X > 0 {
			testCells = append(testCells, &board[myhead.Y+1][myhead.X-1])
		}
		if myhead.Y < req.Height-2 {
			testCells = append(testCells, &board[myhead.Y+2][myhead.X])
		}
		if myhead.X < req.Width-1 {
			testCells = append(testCells, &board[myhead.Y+1][myhead.X+1])
		}
	} else if dir == LEFT && myhead.X > 0 && board[myhead.Y][myhead.X-1].t != SNAKE {
		if myhead.Y > 0 {
			testCells = append(testCells, &board[myhead.Y-1][myhead.X-1])
		}
		if myhead.X > 1 {
			testCells = append(testCells, &board[myhead.Y][myhead.X-2])
		}
		if myhead.Y < req.Height-1 {
			testCells = append(testCells, &board[myhead.Y+1][myhead.X-1])
		}
	} else if dir == RIGHT && myhead.X < req.Width-1 && board[myhead.Y][myhead.X+1].t != SNAKE {
		if myhead.Y > 0 {
			testCells = append(testCells, &board[myhead.Y-1][myhead.X+1])
		}
		if myhead.X < req.Width-2 {
			testCells = append(testCells, &board[myhead.Y][myhead.X+2])
		}
		if myhead.Y < req.Height-1 {
			testCells = append(testCells, &board[myhead.Y+1][myhead.X+1])
		}
	} else {
		return 0
	}

	all_tests := true
	possible := false

	for _, cell := range testCells {
		if cell.t != SNAKE || cell.pos == len(getSnake(req, cell.snake).Coords)-1 {
			all_tests = false
		}
		if cell.t == SNAKE && cell.pos == 0 && mylen <= len(getSnake(req, cell.snake).Coords) {
			possible = true
		}
	}

	if all_tests {
		return 0
	} else if possible {
		return 1
	} else {
		return 2
	}
}

func firstSafeDir(data *TurnData) Dir {
	var dir Dir
	var risky Dir
	for dir = UP; dir < num_dirs; dir++ {
		safety := safeMove(data, dir)
		if safety == 2 {
			return dir
		} else if safety == 1 {
			risky = dir
		}
	}

	return risky
}

func findFood(data *TurnData) Dir {
	shortest := Point{-1, -1}
	food_list := data.req.Food
	myhead := data.mysnake.Coords[0]
	short_dist := -1

	if len(food_list) == 0 {
		return firstSafeDir(data)
	}

	for _, food := range food_list {
		dist := food.X + food.Y - (myhead.X + myhead.Y)
		if dist < short_dist || short_dist == -1 {
			shortest = food
			short_dist = dist
		}
	}

	x_dist := myhead.X - shortest.X
	y_dist := myhead.Y - shortest.Y

	risky := Dir(-1)

	if x_dist < 0 { // go right
		safety := safeMove(data, RIGHT)
		if safety == 2 {
			return RIGHT
		} else if safety == 1 && risky < 0 {
			risky = RIGHT
		}
	}
	if x_dist > 0 { // go left
		safety := safeMove(data, LEFT)
		if safety == 2 {
			return LEFT
		} else if safety == 1 && risky < 0 {
			risky = LEFT
		}
	}
	if y_dist < 0 { //go down
		safety := safeMove(data, DOWN)
		if safety == 2 {
			return DOWN
		} else if safety == 1 && risky < 0 {
			risky = DOWN
		}
	}
	if y_dist < 0 { //go down
		safety := safeMove(data, UP)
		if safety == 2 {
			return UP
		} else if safety == 1 && risky < 0 {
			risky = UP
		}
	}

	if risky < 0 {
		return firstSafeDir(data)
	} else {
		return risky
	}

}

func respond(res http.ResponseWriter, obj interface{}) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(obj)
}

func handleStart(res http.ResponseWriter, req *http.Request) {
	_, err := NewGameStartRequest(req)
	if err != nil {
		respond(res, GameStartResponse{
			Taunt:   toStringPointer("Whoa dude"),
			Color:   "#00FF00",
			Name:    "Skate Fast Eat Gushers",
			HeadUrl: toStringPointer(fmt.Sprintf("%v://%v/static/head.png")),
		})
	}

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	respond(res, GameStartResponse{
		Taunt:   toStringPointer("Whoa dude"),
		Color:   "#00FF00",
		Name:    "Skate Fast Eat Gushers",
		HeadUrl: toStringPointer(fmt.Sprintf("%v://%v/static/head.png", scheme, req.Host)),
	})
}

func handleMove(res http.ResponseWriter, req *http.Request) {
	data, err := NewMoveRequest(req)
	if err != nil {
		respond(res, MoveResponse{
			Move:  "up",
			Taunt: toStringPointer("can't parse this!"),
		})
		return
	}

	snake := getSnake(data, data.You)
	turnData := &TurnData{req: data, board: buildBoard(data), mysnake: &snake}

	dir := findFood(turnData)

	respond(res, MoveResponse{
		Move:  directions[dir],
		Taunt: &data.You,
	})
}
