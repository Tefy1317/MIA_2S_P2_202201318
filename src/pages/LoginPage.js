import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './LoginPage.css';

const LoginPage = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [id, setId] = useState('');
  const navigate = useNavigate();

  const handleLogin = () => {
    const command = `login -user=${username} -pass=${password} -id=${id}`;
    fetch('http://localhost:3001/execute', { 
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ commands: command }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.result && data.result.includes("Inicio de sesión exitoso")) {
          localStorage.setItem('loggedInUser', username);
          alert('Inicio de sesión exitoso');
          navigate('/home'); 
        } else {
          alert('Credenciales incorrectas');
        }
      })
      .catch((error) => {
        console.error('Error en la solicitud de inicio de sesión:', error);
      });
  };

  return (
    <div className="login-container">
      <h2>Iniciar Sesión</h2>
      <div className="form-group">
        <label>Usuario:</label>
        <input
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </div>
      <div className="form-group">
        <label>Contraseña:</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </div>
      <div className="form-group">
        <label>ID Partición:</label>
        <input
          type="text"
          value={id}
          onChange={(e) => setId(e.target.value)}
        />
      </div>
      <button onClick={handleLogin}>Iniciar Sesión</button>
    </div>
  );
};

export default LoginPage;
