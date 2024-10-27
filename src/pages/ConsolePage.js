import React, { useState, useEffect } from 'react';
import InputArea from '../components/InputArea'; 
import OutputArea from '../components/OutputArea';
import { useNavigate } from 'react-router-dom';

const ConsolePage = () => {
  const [output, setOutput] = useState('');
  const [loggedInUser, setLoggedInUser] = useState(null); 
  const navigate = useNavigate();

  useEffect(() => {
    const user = localStorage.getItem('loggedInUser');
    if (user) {
      setLoggedInUser(user); 
    }
  }, []);

  const handleLogout = () => {
    const command = `logout`;
    fetch('http://localhost:3001/execute', { 
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ commands: command }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.result && data.result.includes("Sesión cerrada exitosamente")) {
          alert('Sesión cerrada exitosamente');
          localStorage.removeItem('loggedInUser'); 
          setLoggedInUser(null);
          navigate('/login');
        } else {
          alert('Credenciales incorrectas');
        }
      })
      .catch((error) => {
        console.error('Error en la solicitud de inicio de sesión:', error);
      });
  };

  const handleExecute = (result) => {
    setOutput(result);
  };

  return (
    <div>
      <header>
        {loggedInUser ? (
          <div className="user-info">
            <span className="user-label">{`Usuario: ${loggedInUser}`}</span>
            <button className="logout-button" onClick={handleLogout}>
              Cerrar Sesión
            </button>
          </div>
        ) : (
          <div className="user-info">
            <button className="login-button" onClick={() => navigate('/login')}>
              Iniciar Sesión
            </button>
          </div>
        )}
      </header>
      
      <div className="consola-container">
        <h2>Entrada</h2>
        <InputArea onExecute={handleExecute} />
        <OutputArea output={output} />
      </div>
    </div>
  );
};

export default ConsolePage;