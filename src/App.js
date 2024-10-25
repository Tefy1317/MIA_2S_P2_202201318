import React from 'react';
import { BrowserRouter as Router, Route, Routes, Link } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import ConsolePage from './pages/ConsolePage';
import './App.css';

function App() {
  return (
    <Router>
      <div>
        <nav className="navbar">
          <h1>Proyecto 2 - MIA</h1>
          <ul>
            <li>
              <Link to="/home">Home</Link>
            </li>
            <li>
              <Link to="/login">Iniciar sesión</Link>
            </li>
            <li>
              <Link to="/visualizador">Visualizador</Link>
            </li>
          </ul>
        </nav>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/home" element={<ConsolePage />} />
          <Route path="/" element={<ConsolePage />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
