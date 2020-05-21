let hexColor = function(c) {
  if (c < 256) {
    let x = Math.floor(Math.abs(c)).toString(16);
    if (x.length < 2) {
      return '0' + x;
    }
    return x;
  }
  return '00';
};

let onload = () => {
  let table = document.getElementById('table');
  const L = 16;
  for (let i = 0; i < L; i++ ) {
    let tr = document.createElement('TR');
    for (let j = 0; j < L; j++) {
      let td = document.createElement('TD');
      let color = '#ff' + hexColor(255*(i+1)/L) + hexColor(255*(j+1)/L);
      td.style.backgroundColor = color;
      td.appendChild(document.createTextNode(' '));
      tr.appendChild(td);
    }
    table.appendChild(tr);
  }
  let tableChangeColor = function(e) {
    if (e.target.tagName == 'TD') {
      console.log(e.target);
    }
  };
  table.addEventListener('mouseover', tableChangeColor, false);
};

let draw_started = new Date();
let draw_times = [];
let draw = function() {
  let now = new Date();
  draw_times.push(now);
  if (draw_times.length > 10) {
    draw_times.shift();
  }
  if (draw_times.length > 1) {
    let oldest = draw_times[0];
    let newest = draw_times[draw_times.length-1];
    let elapsed = (newest - oldest)/1000; // seconds
    let fps = (draw_times.length-1)/elapsed;
    let span = document.getElementById('fps');
    span.replaceChild(document.createTextNode(''+fps), span.lastChild);
  }
  if ((now - draw_started)/1000 < 20) {
    requestAnimationFrame(draw);
  }
}
draw();

window.addEventListener('load', (event) => {
  let now = new Date();
  console.log('page is fully loaded' + now.toString());
  onload();
});
