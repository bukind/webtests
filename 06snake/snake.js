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

// getXY returns the cell at (X,Y).
function getXY(x, y) {
  let id = "Y" + y + "X" + x;
  return document.querySelector("#" + id);
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
      td.id = "Y" + i + "X" + j;
      tr.appendChild(td);
    }
    table.appendChild(tr);
  }
  getXY(game.x, game.y).className = "head";
  game.placeFood();
  document.addEventListener('keydown', e => {
    switch (e.keyCode) {
    case 37: // arrows follow
    case 38:
    case 39:
    case 40:
    case 32: // space
    case 27: // escape
    case 65: // wsad follow
    case 87:
    case 83:
    case 68:
    case 89: // autopilot
      game.events.push(e.keyCode);
      break;
    default:
      console.log("keycode = " + e.keyCode);
      return;
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
  let e = null;
  while (game.events.length > 0) {
    e = game.events.shift()
    let dx = game.dx;
    let dy = game.dy;
    switch (e) {
    case 27: // end of game
      return false;
    case 32: // pause
      dx = 0;
      dy = 0;
      break;
    case 89: // Y, autopilot
      game.autopilot = !game.autopilot;
      if (!game.autopilot) {
        dx = 0;
        dy = 0;
        break;
      }
    case 37: // <-
    case 65: // A
      dx = -1;
      dy = 0;
      break;
    case 38: // ^
    case 87: // W
      dx = 0;
      dy = -1;
      break;
    case 39: // ->
    case 68: // D
      dx = 1;
      dy = 0;
      break;
    case 40: // v
    case 83: // S
      dx = 0;
      dy = 1;
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
function moveHead(x, y) {
  let game = this;
  let head = getXY(x, y);
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
  return true;
}


// moveSnake is called every tick to move the snake.
function moveSnake() {
  let game = this;
  if (!game.parseInput()) {
    game.stop();
    return;
  }
  const [x, y] = normXY(game.x + game.dx, game.y + game.dy);
  if (x != game.x || y != game.y) {
    // Sdvig bashki.
    if (!game.moveHead(x, y)) {
      return;
    }
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
