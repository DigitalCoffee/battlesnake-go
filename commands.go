package main

import (
	"encoding/json"
	//astar "github.com/beefsack/go-astar"
	"fmt"
	"log"
	"net/http"
	"time"
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

/*type Path struct {
	p             Point
	g             [][]Cell
	height, width int
}

func (p *Path) PathNeighbors() []astar.Pather {
	neighbors := make([]astar.Pather, 0, 4)
	if p.p.X > 0 && p.g[p.p.Y][p.p.X-1].t != SNAKE {
		neighbors = append(neighbors, &Path{p: Point{X: p.p.X - 1, Y: p.p.Y}, g: p.g, height: p.height, width: p.width})
	}
	if p.p.X < p.width-1 && p.g[p.p.Y][p.p.X+1].t != SNAKE {
		neighbors = append(neighbors, &Path{p: Point{X: p.p.X + 1, Y: p.p.Y}, g: p.g, height: p.height, width: p.width})
	}
	if p.p.Y > 0 && p.g[p.p.Y-1][p.p.X].t != SNAKE {
		neighbors = append(neighbors, &Path{p: Point{X: p.p.X, Y: p.p.Y - 1}, g: p.g, height: p.height, width: p.width})
	}
	if p.p.Y > p.height-1 && p.g[p.p.Y+1][p.p.X].t != SNAKE {
		neighbors = append(neighbors, &Path{p: Point{X: p.p.X, Y: p.p.Y + 1}, g: p.g, height: p.height, width: p.width})
	}
	return neighbors
}

func (p *Path) PathNeighborCost(to astar.Pather) float64 {
	return 1
}
func (p *Path) PathEstimatedCost(to astar.Pather) float64 {
	return 1
}

/*
func AStar(head Point, go_to Point, data *TurnData) (Point, bool) {
	start := &Path{p: head, g: data.board, height: data.req.Height, width: data.req.Width}
	end := &Path{p: go_to, g: data.board, height: data.req.Height, width: data.req.Width}
	path, _, found := astar.Path(start, end)
	if !found {
		return Point{-1, -1}, false
	}
	return path[0].(*Path).p, true
}*/

/*
func AStar(head Point, go_to Point, data TurnData) Point {
	// Reference: https://en.wikipedia.org/wiki/A*_search_algorithm#Pseudocode
	//Setup
	width := data.req.Width
	height := data.req.Height
	closedSet := []Point{}
	openSet := []Point{head}
	cameFrom := []Point{}
	// cost to get to individual point
	// access point as go_score[i][j] = go_score[i*width + j]
	go_score := make([]int, height*width)
	go_score[head.Y*width+head.X] = 0
	//cost of travelling to go_to through the point
	through_score := make([]int, height*width)
	through_score[head.y*width+head.x] = heuristic_cost(head, go_to)
	// Computation
	for len(openSet) > 0 {


		current := through_score[0] //the node in openSet having the lowest through_score[] value
		if current == goal {
			return reconstruct_path(cameFrom, current)
		}
		openSet = append(openSet[:current], openSet[current+1:])
		closedSet = append(current)
	}
	return Point{-1, -1}
}

func reconstruct_path(from []Point, current Point) Point {
	return Point{-1, -1}
}*/

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

type Path struct {
	Point
	prev *Path
	dir  Dir
}

func (p *Path) Len() int {
	if p.prev == nil {
		return 0
	}
	return p.prev.Len() + 1
}

func (p *Path) pointInPath(point Point) bool {
	if p.Point == point {
		return true
	} else if p.prev == nil {
		return false
	} else {
		return p.prev.pointInPath(point)
	}
}

func cell(board [][]Cell, p Point) *Cell {
	return &board[p.Y][p.X]
}

func bfs(data *TurnData, attack bool) Dir {
	req := data.req
	board := data.board

	queue := make([]Path, 1, 10)
	queue[0] = Path{Point: data.mysnake.Coords[0], prev: nil, dir: -1}

	for len(queue) > 0 {
		path := &queue[0]

		if path.Len() > len(data.mysnake.Coords) {
			found := false
			for ; path.prev.dir != -1; path = path.prev {
				c := cell(board, path.Point)
				if attack {
					if c.t == SNAKE && c.snake != data.mysnake.Id && c.pos < path.Len() {
						found = true
					}
				} else {
					if c.t == FOOD {
						found = true
					}
				}
			}
			if found {
				return path.dir
			}
		}
		queue = queue[1:]

		test := Point{X: path.X, Y: path.Y - 1}
		if path.Y > 0 && (cell(board, test).t != SNAKE || cell(board, test).pos < path.Len()) && path.dir != DOWN && !path.pointInPath(test) {
			queue = append(queue, Path{Point: test, prev: path, dir: UP})
		}
		test = Point{X: path.X, Y: path.Y + 1}
		if path.Y < req.Height-1 && (cell(board, test).t != SNAKE || cell(board, test).pos < path.Len()) && path.dir != UP && !path.pointInPath(test) {
			queue = append(queue, Path{Point: test, prev: path, dir: DOWN})
		}
		test = Point{X: path.X - 1, Y: path.Y}
		if path.X > 0 && (cell(board, test).t != SNAKE || cell(board, test).pos < path.Len()) && path.dir != RIGHT && !path.pointInPath(test) {
			queue = append(queue, Path{Point: test, prev: path, dir: LEFT})
		}
		test = Point{X: path.X + 1, Y: path.Y}
		if path.X < req.Width-1 && (cell(board, test).t != SNAKE || cell(board, test).pos < path.Len()) && path.dir != LEFT && !path.pointInPath(test) {
			queue = append(queue, Path{Point: test, prev: path, dir: RIGHT})
		}

	}
	return UP
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

func firstSafeDir(data *TurnData) (Dir, int) {
	var dir Dir
	var risky Dir
	for dir = UP; dir < num_dirs; dir++ {
		safety := safeMove(data, dir)
		if safety == 2 {
			return dir, 2
		} else if safety == 1 {
			risky = dir
		}
	}

	return risky, 1
}

func target(data *TurnData, t Point) Dir {
	myhead := data.mysnake.Coords[0]
	x_dist := myhead.X - t.X
	y_dist := myhead.Y - t.Y

	risky := Dir(-1)

	if abs(x_dist) > abs(y_dist) {
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
	} else {
		if y_dist < 0 { //go down
			safety := safeMove(data, DOWN)
			if safety == 2 {
				return DOWN
			} else if safety == 1 && risky < 0 {
				risky = DOWN
			}
		}
		if y_dist > 0 { //go up
			safety := safeMove(data, UP)
			if safety == 2 {
				return UP
			} else if safety == 1 && risky < 0 {
				risky = UP
			}
		}
	}

	dir, safety := firstSafeDir(data)
	if risky >= 0 && safety == 1 {
		return risky
	} else {
		return dir
	} /*

		p, found := AStar(myhead, shortest, data)

		if !found {
			return firstSafeDir(data)
		} else if p.X == myhead.X-1 {
			return LEFT
		} else if p.X == myhead.X+1 {
			return RIGHT
		} else if p.Y == myhead.Y-1 {
			return UP
		} else if p.Y == myhead.Y+1 {
			return DOWN
		}
		return firstSafeDir(data)*/
}

/*
func findFood(data *TurnData) Dir {
	shortest := Point{-1, -1}
	food_list := data.req.Food
	myhead := data.mysnake.Coords[0]
	short_dist := -1

	if len(food_list) == 0 {
		dir, _ := firstSafeDir(data)
		return dir
	}

	for _, food := range food_list {
		dist := heuristic_cost(myhead, food)
		if dist < short_dist || short_dist == -1 {
			shortest = food
			short_dist = dist
		}
	}

	return target(data, shortest)
}*/

func findEnemy(data *TurnData) Dir {
	shortest := Point{-1, -1}
	snake_list := data.req.Snakes
	myhead := data.mysnake.Coords[0]
	short_dist := -1

	for _, snake := range snake_list {
		if snake.Id == data.mysnake.Id {
			continue
		}
		dist := heuristic_cost(myhead, snake.Coords[0])
		if dist < short_dist || short_dist == -1 {
			shortest = snake.Coords[0]
			short_dist = dist
		}
	}

	return target(data, shortest)
}

func respond(res http.ResponseWriter, obj interface{}) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(obj)
}

func handleStart(res http.ResponseWriter, req *http.Request) {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	respond(res, GameStartResponse{
		Taunt:      toStringPointer("Whoa dude"),
		Color:      "#00FF00",
		Name:       "Skate Fast Eat Gushers",
		Head:       "shades",
		Head_Image: toStringPointer(fmt.Sprintf("%v://%v/head.png", scheme, req.Host)),
	})
}

func handleMove(res http.ResponseWriter, req *http.Request) {
	timer := time.Now()
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

	attack := false
	if snake.HealthPoints > 25 {
		attack = true
		for _, s := range data.Snakes {
			if s.Id != data.You && len(s.Coords) >= len(snake.Coords) {
				attack = false
			}
		}
	}
	var dir Dir
	/*if attack {
		dir = findEnemy(turnData)
	} else {
		dir = findFood(turnData)
	}*/
	dir = bfs(turnData, attack)

	respond(res, MoveResponse{
		Move:  directions[dir],
		Taunt: &data.You,
	})
	t := time.Since(timer)
	if t >= 200*time.Millisecond {
		log.Println("Timed out: ", t)
	}
}
