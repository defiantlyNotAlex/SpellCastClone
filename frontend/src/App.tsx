import { useState, useEffect } from 'react'
import './App.css'

const alphabet = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"];
const scores = [1, 4, 5, 3, 1, 5, 3, 4, 1, 7, 6, 3, 4, 2, 1, 4, 8, 2, 2, 2, 4, 5, 5, 7, 4, 8];

class Position {
  x: number = 0;
  y: number = 0;

  constructor (x: number, y: number) {
    this.x = x;
    this.y = y;
  }

}


const copy_board = <T,>(board: T[][]): T[][] => {
  return [[...board[0]], [...board[1]], [...board[2]], [...board[3]], [...board[4]], [...board[5]]]
}

const false_board = [[false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false], [false, false, false, false, false, false]];

function App() {
  const [board, setBoard] = useState<number[][]>([[0, 1, 2, 3, 4, 5], [6, 7, 8, 9, 10, 11], [12, 13, 14, 15, 16, 17], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0], [0, 0, 0, 0, 0, 0]])
  const [doubleLetter, setDoubleLetter] = useState<{x: number, y: number}>({x: 0, y: 0})
  const [doubleWord, setDoubleWord] = useState<{x: number, y: number}>({x: -1, y: -1})
  const [doubleLetterMult, setDoubleLetterMult] = useState(2)
  const [gemPositions, setGemPositions] = useState<{x: number, y: number}[]>([])

  const [current_word, setCurrentWord] = useState("")
  const [used, setUsed] = useState<boolean[][]>(false_board)
  const [last_pos, setLastPos] = useState<Position | undefined>()
  const [path, setPath] = useState<Position[]>([])
  const [_ws, setWs] = useState<WebSocket | null>(null);

  const [gameData, setGameData] = useState<{turn: number, gameTurn: number, wordsPlayed: {word: string, score: number, playerName: string}[]}>({turn: 0, gameTurn: 0, wordsPlayed: []})
  const [myTurn, setMyTurn] = useState<number>(0)
  const [players, setPlayers] = useState<{name: string, score: number, gems: number}[]>([])

  const [swapX, setSwapX] = useState(0)
  const [swapY, setSwapY] = useState(0)
  const [letter, setLetter] = useState(0)

  const [name, setName] = useState("")

  useEffect(() => {
    fetch('http://localhost:8080/session', {})

    const websocket = new WebSocket('ws://localhost:8080/join');
    setWs(websocket);

    websocket.onopen = () => console.log('Connected to WebSocket server');
    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      const game = data.gameInfo;
      console.log(data);

      setMyTurn(data.yourTurn);

      setBoard(game.boardInfo.board);
      setDoubleLetter(game.boardInfo.doubleLetter);
      setDoubleWord(game.boardInfo.doubleWord);
      setDoubleLetterMult(game.boardInfo.doubleLetterMult);
      setGemPositions(game.boardInfo.gemPositions);

      setPlayers(game.playerInfo.players);
      setGameData(game.gameInfo);
    };
    websocket.onclose = () => console.log('Disconnected from WebSocket server');

    // Cleanup on unmount
    return () => websocket.close();
  }, []);

  const resetBoard = () => {
    setLastPos(undefined);
    setCurrentWord("");
    setPath([]);
    setUsed(false_board);
  }

  return (
    <>
      <div className="gameRoot">
        <div className="left">
          <div style={{width: 40 * 6, height: 40}}>
            <p>{current_word}</p>
          </div>
          <table>
            {board.map((row, y) =>
              <tr>
                {row.map((val, x) =>
                  <td>
                      <button 
                        className='cell'
                        style={{
                          outlineColor: used[y][x] ? "red" : "black",
                          outlineWidth: 2,
                          outlineStyle: 'solid',
                          width: 60,
                          height: 60, 
                          backgroundColor: (x == doubleWord.x && y == doubleWord.y) ? "lightblue" : undefined,
                        }}
                        disabled={
                          used[y][x] 
                          || (last_pos !== undefined ? !(Math.abs(last_pos.x - x) <= 1 && Math.abs(last_pos.y - y) <= 1)  : false)
                          || gameData.turn != myTurn
                        }
                        onClick={() => {
                          setCurrentWord(current_word + alphabet[val]);
                          setPath([...path, new Position(x, y)])
                          var used2 = copy_board(used);
                          used2[y][x] = true;
                          setUsed(used2)
                          setLastPos(new Position(x, y));
                        }}>
                        {alphabet[val]}
                        <br/>
                        <span className='gem'>{gemPositions.find((p, _i, _obj) => p.x == x && p.y == y) !== undefined ? "(G)" : ""}</span>
                        <span className='score' style={{color: (x == doubleLetter.x && y == doubleLetter.y) ? doubleLetterMult === 2 ? "blue" : "red" : undefined}}>
                          [{scores[letter] * ((x == doubleLetter.x && y == doubleLetter.y) ? doubleLetterMult : 1)}]
                        </span>
                        
                      </button>
                  </td>
                )}
              </tr>
            )}
          </table>
          <button onClick={() => {
            console.log(JSON.stringify({path: path}))
            fetch('/turn', {
              method: "POST",
              body: JSON.stringify({path: path}),
            })
            .then(response => response.json())
            .then(json => {console.log(json); if (!json.isWord) {alert("Not In Word List");}})
            .catch(error => console.error(error));

            resetBoard()
          }}>submit</button>
          <button onClick={() => {
            resetBoard()
          }}>clear</button>
          <button onClick={() => {fetch('/shuffle', {})
            .then(response => {
              response.json().then(json => {if (json.message !== undefined) {alert(json.message)}})
            }).catch(error => console.error(error));
            resetBoard();
          }}>
            shuffle
          </button>
          <br/>
          <input style={{width: 40}} onChange={e => setSwapX(parseInt(e.target.value))} type="number" defaultValue={0} max={5}/>
          <input style={{width: 40}} onChange={e => setSwapY(parseInt(e.target.value))} type="number" defaultValue={0} max={5}/>
          <select style={{width: 40}} onChange={e => setLetter(parseInt(e.target.value))} defaultValue={0}>
            {alphabet.map((c, i) => 
              <option value={i}>
                {c}
              </option>
            )}
          </select>
          <button onClick={() => {
            fetch('/swap', {method: "POST", body: JSON.stringify({position: {x: swapX, y: swapY}, letter: letter})})
            .then(response => {
              response.json().then(json => {if (json.message !== undefined) {alert(json.message)}})
            }).catch(error => console.error(error));
          }}>swap</button>
        </div>
        <div className="right">
          <input onChange={e => setName(e.target.value)}/>
          <button onClick={() => {
              fetch('setName', {method: "POST", body: JSON.stringify({name: name})}).catch(error => console.error(error))
          }}>change name</button>
          {players.map((player, i) => 
            <li><p style={{color: gameData.turn == i ? "red" : undefined}}>{player.name}: score {player.score}, gems: {player.gems}</p></li>
          )}
        </div>
        <div className="wordsList">
          {gameData.wordsPlayed.map((row, _i) => 
            <li><p>{row.playerName} : {row.word} | {row.score}</p></li>
          )}
        </div>
      </div>
    </>
  )
}

export default App
