import React, { useState } from 'react';

const InputArea = ({ onExecute }) => {
  const [input, setInput] = useState('');

  const handleInputChange = (e) => {
    setInput(e.target.value);
  };

  const handleExecute = () => {
    fetch('http://localhost:3001/execute', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ commands: input }), 
    })
      .then((response) => response.json())
      .then((data) => {
        console.log('Respuesta del backend:', data); 
        onExecute(data.result); 
      })
      .catch((error) => {
        console.error('Error:', error);
      });
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const fileContent = event.target.result;
        console.log(fileContent); 
        setInput(fileContent); 
      };
      reader.readAsText(file); 
    }
  };

  return (
    <div>
      <textarea
        className="textarea-input"
        value={input}
        onChange={handleInputChange}
        placeholder="Ingresa los comandos aquí..."
        rows="10"
        cols="50"
        style={{ width: '100%', marginBottom: '10px' }}
      />
      <div className="button-container">
        <button className="button" onClick={handleExecute}>
          Ejecutar
        </button>
        <input
          type="file"
          className="input-file"
          accept=".smia"
          onChange={handleFileChange}
        />
      </div>
    </div>
  );
};

export default InputArea;