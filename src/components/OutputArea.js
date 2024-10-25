import React from 'react';

const OutputArea = ({ output }) => {
  return (
    <textarea
      className="output-area"
      value={output}
      readOnly
      rows="10"
      cols="50"
      style={{ width: '100%', marginTop: '10px' }}
      placeholder="Aquí se mostrarán los resultados..."
    />
  );
};

export default OutputArea;