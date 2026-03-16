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
  split(net, left, mid, right) { console.log(JSON.stringify({v:1,type:"split",net,left,mid,right})); }
  merge(net, left, mid, right) { console.log(JSON.stringify({v:1,type:"merge",net,left,mid,right})); }
  write(net, pos, value) {
    const s = this._state[net] || [];
    if (pos >= 0 && pos < s.length) s[pos] = value;
    console.log(JSON.stringify({v:1,type:"write",net,pos,value}));
  }
  not_found(net) { this.notFound(net); }
}
module.exports = { Probe };
