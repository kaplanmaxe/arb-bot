import React from 'react';
// import WebSocket from 'ws';
import logo from './logo.svg';
import './App.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import proto from './arb';

const markets = [];

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      markets: {},
      sortedMarkets: [],
    };
  }

  componentDidMount() {
    let url;
    if (process.env.NODE_ENV == 'production') {
      url = 'ws://api-public.helgart.com/arb';
    } else {
      url = 'ws://localhost:8000/arb';
    }
    this.socket = new WebSocket(url);
    this.socket.binaryType = 'arraybuffer'
  
    // Listen for messages
    this.socket.addEventListener('message', this.msgListener);
  }

  // updateMarkets = (market) => {
  //   const stateCopy = this.state.markets.concat();
  //   stateCopy[market.hePair] = market;
  //   this.setState({
  //     markets: stateCopy,
  //   });
  // }

  msgListener =(event) => {
    var message = proto.wsapi.ArbMarket.decode(new Uint8Array(event.data));
    const markets = {...this.state.markets};
    markets[message.hePair] = message;
    const sortedMarkets = [];
    for (let i = 0; i < Object.keys(markets).length; i++) {
      sortedMarkets.push(markets[Object.keys(markets)[i]]);
    }
    sortedMarkets.sort((a, b) => {
      // console.log(a.spread, b.spread);
      return b.spread - a.spread;
    });
    // console.log(sortedMarkets);
    this.setState({
      markets: markets,
      sortedMarkets: sortedMarkets,
    });
  }
  render() {
    // this._foo();
    return (
      <div className="App">
        <table className="table">
          <thead>
            <tr>
              <th scope="col">Pair</th>
              <th scope="col">Spread</th>
              <th scope="col">Low Exchange</th>
              <th scope="col">Low Exchange Price</th>
              <th scope="col">High Exchange</th>
              <th scope="col">High Exchange Price</th>
            </tr>
          </thead>
          <tbody>
            {this.state.sortedMarkets.map(res => {
              const market = res;
              return (
                <tr  key={market.hePair}>
                  <th scope="row">{market.hePair}</th>
                  <td>{Number(market.spread).toFixed(2)}%</td>
                  <td>{market.low.exchange}</td>
                  <td>{market.low.price}</td>
                  <td>{market.high.exchange}</td>
                  <td>{market.high.price}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    );
  }
}
// function App() {
//   return (
//     <div className="App">
//       <header className="App-header">
//         <img src={logo} className="App-logo" alt="logo" />
//         <p>
//           Edit <code>src/App.js</code> and save to reload.
//         </p>
//         <a
//           className="App-link"
//           href="#"
//           onClick={() => foo()}
//           rel="noopener noreferrer"
//         >
//           Learn React
//         </a>
//       </header>
//     </div>
//   );
// }

// function foo() {
//   const socket = new WebSocket('ws://localhost:8000/arb');
//   socket.binaryType = 'arraybuffer'

//   // Listen for messages
//   socket.addEventListener('message', function (event) {
//       var message = proto.wsapi.ArbMarket.decode(new Uint8Array(event.data));
//       markets[message.hePair] = message;
//       console.log(markets);
//       // socket.close();
//   });
// }

export default App;
