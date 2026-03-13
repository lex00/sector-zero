// Minimal JS probe bundled with game binary
class Probe {
  constructor() { this._state = {}; }
  init(net, values) {
    this._state[net] = [...values];
    console.log(JSON.stringify({v:1,type:"init",net,values}));
  }
  compare(net, i, j) { console.log(JSON.stringify({v:1,type:"compare",net,i,j})); }
  swap(net, i, j) {
    const s = this._state[net] || [];
    [s[i], s[j]] = [s[j], s[i]];
    console.log(JSON.stringify({v:1,type:"swap",net,i,j}));
  }
  pin(net, name, pos) { console.log(JSON.stringify({v:1,type:"pin",net,name,pos})); }
  signal(net, name, positions) { console.log(JSON.stringify({v:1,type:"signal",net,name,positions})); }
  access(net, pos) { console.log(JSON.stringify({v:1,type:"access",net,pos})); }
  found(net, pos) { console.log(JSON.stringify({v:1,type:"found",net,pos})); }
  notFound(net) { console.log(JSON.stringify({v:1,type:"not_found",net})); }
  bounds(net, low, high) { console.log(JSON.stringify({v:1,type:"bounds",net,low,high})); }
  done(net) { console.log(JSON.stringify({v:1,type:"done",net})); }
}
module.exports = { Probe };
