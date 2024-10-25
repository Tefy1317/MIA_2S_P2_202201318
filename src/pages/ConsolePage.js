import React, { useState } from 'react';
import InputArea from '../components/InputArea'; 
import OutputArea from '../components/OutputArea';

const ConsolePage = () => {
  const [output, setOutput] = useState('');

  const handleExecute = (result) => {
    setOutput(result);
  };

  return (
    <div className="consola-container">
      <h2>Entrada</h2>
      <InputArea onExecute={handleExecute} />
      <OutputArea output={output} />
    </div>
  );
};

export default ConsolePage;