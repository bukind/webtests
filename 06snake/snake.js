"use strict";

const H = 30;
const W = 40;

// normXY returns x and y, properly normalized.
function normXY(x, y) {
  if (x < 0) {
    x = W-1;
  } else if (x >= W) {
    x = 0;
  }
  if (y < 0) {
    y = H-1;
  } else if (y >= H) {
    y = 0;
  }
  return [x, y];
}

function id(x, y) {
  return "Y" + y + "X" + x;
}

const dirs = {
  "right": [1, 0],
  "top": [0, -1],
  "left": [-1, 0],
  "bottom": [0, 1],
};

// getXY returns the cell at (X,Y).
function getXY(x, y) {
  return document.querySelector("#" + id(x,y));
}

// start inits the game.
function start() {
  let game = this;
  let table = document.getElementById("table");
  let tr = document.createElement("TR");
  let th = document.createElement("TH");
  th.colSpan = W;
  let span = document.createElement("SPAN");
  span.id = "report";
  th.appendChild(span);
  tr.appendChild(th);
  table.appendChild(tr);
  let eog = document.createElement("DIV");
  eog.id = "end";
  table.appendChild(eog);

  game.canvas = table;
  for (let i = 0; i < H; i++) {
    let tr = document.createElement("TR");
    for (let j = 0; j < W; j++) {
      let td = document.createElement("TD");
      td.className = "empty";
      td.id = id(j,i);
      tr.appendChild(td);
    }
    table.appendChild(tr);
  }
  getXY(game.x, game.y).className = "head";
  game.placeFood();
  document.addEventListener('keydown', e => {
    var act;
    switch (e.keyCode) {
    case 32: // space
      act = "pause";
      break;
    case 27: // escape
      act = "stop";
      break;
    case 89: // autopilot
      act = "auto";
      break;
    case 37:
    case 65:
      act = "left";
      break;
    case 38:
    case 87:
      act = "top";
      break;
    case 39:
    case 68:
      act = "right"
      break;
    case 40:
    case 83:
      act = "bottom";
      break;
    default:
      console.log("keycode = " + e.keyCode);
      return;
    }
    if (!game.autopilot || (act in dirs === false)) {
      game.events.push(act);
    }
  }, false);
  game.interval = setInterval(() => {
    game.moveSnake();
  }, 100);
}

// stop stops the game.
function stop() {
  let game = this;
  clearInterval(this.interval);

  let tl = getXY(3, Math.floor(H/2)-2).getBoundingClientRect();
  let br = getXY(W-4, Math.floor(H/2)+1).getBoundingClientRect();
  let top = Math.floor(tl.top);
  let left = Math.floor(tl.left);
  let bottom = Math.ceil(br.bottom);
  let right = Math.ceil(br.right);

  let canvas = document.getElementById("canvas");
  let div = document.createElement("DIV");
  div.className = "end";
  div.style.position = "absolute";
  div.style.top = top + "px";
  div.style.left = left + "px";
  div.style.width = (right-left) + "px";
  div.style.height = (bottom-top) + "px";
  div.style.lineHeight = (bottom-top) + "px";
  canvas.appendChild(div);
  div.textContent = "GAME OVER";
}

// parseInput parses the keyboard input.
function parseInput() {
  let game = this;
  if (game.events.length === 0) {
    return true;
  }
  while (game.events.length > 0) {
    let act = game.events.shift()
    let dx = game.dx;
    let dy = game.dy;
    switch (act) {
    case "stop":
      return false;
    case "pause":
      [dx, dy] = [0, 0];
      break;
    case "auto": // Y, autopilot
      game.autopilot = !game.autopilot;
      if (!game.autopilot) {
        [dx, dy] = [0, 0];
      }
      break;
    case "left":
    case "top":
    case "right":
    case "bottom":
      [dx, dy] = dirs[act];
      break;
    default:
      console.log("could not get here, keycode=" + e.keyCode);
    }
    if (game.dx !== dx || game.dy !== dy) {
      if (game.autopilot) {
        // we don't care about turning keys.
        continue;
      }
      const [x, y] = normXY(game.x + dx, game.y + dy);
      let test = getXY(x,y);
      if (game.snake.length > 0) {
        if (game.snake[0].id === test.id) {
          // We attempted to turn back, try again.
          continue;
        }
      }
      game.dx = dx;
      game.dy = dy;
      return true;
    }
  }
  return true;
}

// getRandomInt returns an int between [0, max).
function getRandomInt(max) {
  return Math.floor(Math.random() * Math.floor(max));
}

// placeFood places a new food.
function placeFood() {
  let game = this;
  if (game.foodx !== -1 && game.foody !== -1) {
    return true;
  }
  for (let i = 0; i < 1000; i++) {
    let x = getRandomInt(W);
    let y = getRandomInt(H);
    let c = getXY(x,y);
    if (c.className === "empty") {
      c.className = "food";
      game.foodx = x;
      game.foody = y;
      return true;
    }
  }
  return false;
}

// moveHead moves head to the new position x, y.
function moveHead() {
  let game = this;
  if (game.dx === 0 && game.dy === 0) {
    return true;
  }
  const [x, y] = normXY(game.x + game.dx, game.y + game.dy);
  let head = getXY(x, y);
  if (game.autopilot) {
    // detect obstacles, or turn into a pile.
  }
  if (head.className === "food") {
    game.len = game.len + 5;
    game.foodx = -1;
    game.foody = -1;
    if (!game.placeFood()) {
      game.stop();
      return false;
    }
  } else if (head.className !== "empty") {
    game.stop();
    return false;
  }
  game.ticks = game.ticks + 1;
  let c = getXY(game.x, game.y);
  c.className = "body";
  if (game.snake.unshift(c) > game.len) {
    let tail = game.snake.pop();
    tail.className = "empty";
  }
  game.x = x;
  game.y = y;
  head.className = "head";
  if (game.autopilot) {
    if (game.x === game.foodx) {
      game.dy = 1;
      game.dx = 0;
    } else if (game.y === game.foody) {
      game.dx = 1;
      game.dy = 0;
    }
  }
  return true;
}


// moveSnake is called every tick to move the snake.
function moveSnake() {
  let game = this;
  if (!game.parseInput()) {
    game.stop();
    return;
  }
  if (!game.moveHead()) {
    return;
  }
  let th = document.getElementById("report");
  let msg = "Length: " + game.len + " ; Ticks: " + game.ticks;
  if (game.autopilot) {
    msg += ", autopilot";
  }
  th.textContent = msg;
}

let game = {
  // Data.
  canvas: null,
  x: 5,
  y: 5,
  dx: 0,
  dy: 0,
  events: [],
  snake: [],
  len: 0,
  ticks: 0,
  autopilot: false,
  foodx: -1,
  foody: -1,

  // Methods.
  start:     start,
  stop:      stop,
  moveSnake: moveSnake,
  moveHead:  moveHead,
  parseInput: parseInput,
  placeFood: placeFood,
}

let onload = () => {
  game.start();
}
