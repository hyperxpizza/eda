import React, { ChangeEvent, useCallback, useEffect, useState } from 'react';
import './App.css';



function App() {
  const [messages, updateMessages] = useState<string[]>([]);
  const [inputValue, setInputValue] = useState<string>('');

  const socket = new WebSocket("ws://127.0.0.1:8888/ws");


  useEffect(() => {
    socket.onopen = () => {
      updateMessages([...messages, "Connected"])
    };

    socket.onmessage = (e) => {
      updateMessages([...messages, e.data])
    }
    
    return () => {
      socket.close()
    }

  }, [])

  const sendMessage = useCallback((e:any) => {
    e.preventDefault()

    socket.send(JSON.stringify({
      message: inputValue
    }))
    
    setInputValue("");

  },[inputValue])

  const handleChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    e.preventDefault();
    //updateMessages([...messages, e.target.value]);
    setInputValue(e.target.value);
  },[])

  return (
    <div className="App">
      <input id="input" type="text" value={inputValue} onChange={handleChange} />
      <button onClick={sendMessage}>Send</button>
      <div>
        {
          messages.map((item: string) => (
            <div>{item}</div>
          ))
        }
      </div>
    </div>
  )

}

export default App;
