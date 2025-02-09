import React, { useEffect, useState } from "react";
import "./App.css";
import Table from "./Table";

const API_URL = "http://localhost:8080/ping_results";

function App() {
  const [pingResults, setPingResults] = useState([]);

  useEffect(() => {
    fetchPingResults();
    const interval = setInterval(fetchPingResults, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchPingResults = async () => {
    try {
      const response = await fetch(API_URL);
      if (response.ok) {
        const data = await response.json();
        // Сортировка по последнему успешному пингу (от свежего к старому)
        const sortedData = data.sort((a, b) =>
        new Date(b.last_successful) - new Date(a.last_successful)
        );
        setPingResults(sortedData);
      } else {
        console.error("Ошибка запроса данных");
      }
    } catch (error) {
      console.error("Ошибка подключения:", error);
    }
  };

  return (
  <div className="App">
  <h1>Ping Results</h1>
  <Table data={pingResults} />
  </div>
  );
}

export default App;
