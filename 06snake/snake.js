let setTD = (td, kl) => {
  td.classList.remove("empty", "filled");
  td.classList.add(kl);
};

let onload = () => {
  let table = document.getElementById("table");
  const H = 30;
  const W = 40;
  for (let i = 0; i < H; i++) {
    let tr = document.createElement("TR");
    tr.id = "Y" + i;
    for (let j = 0; j < W; j++) {
      let td = document.createElement("TD");
      td.className = "X" + j;
      setTD(td, "empty");

      tr.appendChild(td);
    }
    table.appendChild(tr);
  }
  let c = document.querySelector("#Y5 .X9");
  setTD(c, "filled");
};
