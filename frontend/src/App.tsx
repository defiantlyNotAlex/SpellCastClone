import { useState, useEffect } from 'react'
import './App.css'

const alphabet = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"];
//const scores = [1, 4, 5, 3, 1, 5, 3, 4, 1, 7, 6, 3, 4, 2, 1, 4, 8, 2, 2, 2, 4, 5, 5, 7, 4, 8];

class Position {
  x: number = 0;
  y: number = 0;

  constructor (x: number, y: number) {
    this.x = x;
    this.y = y;
  }

}


// class Game {
//   board: number[][] = [[0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0]];
//   double_letter: Position | undefined;
//   triple_letter: Position | undefined;
//   double_word: Position | undefined;
  
//   current_word: string = "";
//   current_path: Path = new Path;
//   used_tiles: boolean[][] = [[false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false]];
  
//   turn: number = 0;
// }

// class Path {
//   path: Position[] = [];
// }


const copy_board = <T,>(board: T[][]): T[][] => {
  return [[...board[0]], [...board[1]], [...board[2]], [...board[3]], [...board[4]], [...board[5]]]
}

// const update_board = <T,>(board: T[][]): T[][] => {
//   return [];
// }

const board_from_json = (board: {board: number[][]}): number[][] => {
  return board.board.map((row, _y) => {
    return row.map((val, _x) => {
      return val;
    });
  });
}

const false_board = [[false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false]];

function App() {
  const [board, setBoard] = useState<number[][]>([[0, 1, 2, 3, 4, 5], [6, 7, 8, 9, 10, 11], [12, 13, 14, 15, 16, 17], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0]])
  const [_doubleLetter, setDoubleLetter] = useState<number[] | null>([0, 0])
  const [_doubleWord, setDoubleWord] = useState<number[] | null>(null)
  const [current_word, setCurrentWord] = useState("")
  const [used, setUsed] = useState<boolean[][]>(false_board)
  const [last_pos, setLastPos] = useState<Position | undefined>()
  const [path, setPath] = useState<Position[]>([])
  const [_ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    fetch('http://localhost:8080/board')
      .then(response => response.json())
      .then(json => {console.log(json); return json})
      .then(json => setBoard(board_from_json(json)))
      .catch(error => console.error(error));
  }, []);


  useEffect(() => {
    const websocket = new WebSocket('http://localhost:8080/join');
    setWs(websocket);

    websocket.onopen = () => console.log('Connected to WebSocket server');
    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data)
      console.log(data)

      setBoard(data.boardInfo.board)
      setDoubleLetter(data.boardInfo.doubleLetter)
      setDoubleWord(data.boardInfo.doubleWord)
    };
    websocket.onclose = () => console.log('Disconnected from WebSocket server');

    // Cleanup on unmount
    return () => websocket.close();
  }, []);

  return (
    <>
      <div style={{width: 40 * 6, height: 40}}>
        <p>{current_word}</p>
      </div>
      <table>
        {board.map((row, y) =>
          <tr>
            {row.map((val, x) =>
              <td>
                  <button 
                    style={{
                      outlineColor: used[y][x] ? "red" : "black",
                      outlineWidth: 2,
                      outlineStyle: 'solid',
                      width: 40,
                      height: 40,
                    }}
                    disabled={
                      used[y][x] || (last_pos !== undefined ? !(Math.abs(last_pos.x - x) <= 1 && Math.abs(last_pos.y - y) <= 1)  : false)
                    }
                    onClick={() => {
                      setCurrentWord(current_word + alphabet[val]);
                      setPath([...path, new Position(x, y)])
                      var used2 = copy_board(used);
                      used2[y][x] = true;
                      setUsed(used2)
                      setLastPos(new Position(x, y));
                    }}
                  >
                    {alphabet[val]}
                  </button>
              </td>
            )}
          </tr>
        )}
      </table>
      <button onClick={() => {
        console.log(JSON.stringify({path: path}))
        fetch('http://localhost:8080/turn', {
          method: "POST",
          body: JSON.stringify({path: path}),
        })
        .then(response => response.json())
        .then(json => console.log(json))
        .catch(error => console.error(error))
      }}>submit</button>
      <button onClick={() => {
        setLastPos(undefined);
        setCurrentWord("");
        setPath([]);
        setUsed(false_board);
      }}>clear</button>
    </>
  )
}

export default App
