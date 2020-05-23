"use strict";

const H = 30;
const W = 40;

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
  game.getXY(game.x, game.y).className = "head";
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
      game.onKeyDown(e);
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

function stop() {
  let game = this;
  clearInterval(this.interval);

  let tl = game.getXY(3, Math.floor(H/2)-2).getBoundingClientRect();
  let br = game.getXY(W-4, Math.floor(H/2)+1).getBoundingClientRect();
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

function getXY(x, y) {
  let id = "Y" + y + "X" + x;
  return document.querySelector("#" + id);
}

function onKeyDown(e) {
  let game = this;
  game.events.push(e.keyCode);
}

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
      const [x, y] = game.moveHead(dx, dy);
      let test = game.getXY(x,y);
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

function getRandomInt(max) {
  return Math.floor(Math.random() * Math.floor(max));
}

function placeFood() {
  let game = this;
  while (true) {
    let x = getRandomInt(W);
    let y = getRandomInt(H);
    let c = game.getXY(x,y);
    if (c.className === "empty") {
      c.className = "food";
      return;
    }
  }
}

// return new x and y.
function moveHead(dx, dy) {
  let game = this;
  let x = game.x + dx;
  let y = game.y + dy;
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

function moveSnake() {
  let game = this;
  if (!game.parseInput()) {
    game.stop();
    return;
  }
  const [x, y] = game.moveHead(game.dx, game.dy);
  let head = game.getXY(x, y);
  if (x != game.x || y != game.y) {
    // Sdvig bashki.
    if (head.className === "food") {
      game.len = game.len + 5;
      game.placeFood();
    } else if (head.className !== "empty") {
      game.stop();
      return;
    }
    game.ticks = game.ticks + 1;
    let c = game.getXY(game.x, game.y);
    c.className = "body";
    if (game.snake.unshift(c) > game.len) {
      let tail = game.snake.pop();
      tail.className = "empty";
    }
  }
  game.x = x;
  game.y = y;
  head.className = "head";
  let th = document.getElementById("report");
  th.textContent = "Length: " + game.len + " ; Ticks: " + game.ticks;
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
  // Methods.
  start:     start,
  stop:      stop,
  onKeyDown: onKeyDown,
  getXY:     getXY,
  moveSnake: moveSnake,
  parseInput: parseInput,
  placeFood: placeFood,
  moveHead:  moveHead,
}

let onload = () => {
  game.start();
}
