"use client"
import { useState, FormEvent, Dispatch, SetStateAction } from 'react';
import Insights from './insights';

export default function Home() {

  const [name, setName] = useState('');
  const [isLoading, setIsLoading] = useState<boolean>(false)
  const [error, setError] = useState(null);
  const [data, setData] = useState(null);

  async function onSubmit(event: FormEvent<HTMLFormElement>, setData: Dispatch<SetStateAction<any>>) {
    event.preventDefault()
    setIsLoading(true)
    setError(null) // Clear previous errors when a new request starts
 
    try {
      const formData = new FormData(event.currentTarget)
      const name = formData.get("name")
      const response = await fetch(`https://commitcrunch.fly.dev/summary/${name}`, {
        method: 'GET',
      })
 
      if (!response.ok) {
        throw new Error('Failed to submit the data. Please try again.')
      }
 
      // Handle response if necessary
      const js = await response.json()
      setData(js);
    } catch (error: unknown) {
      // Capture the error message to display to the user
      // setError(error.message)
      console.error(error)
    } finally {
      setIsLoading(false)
    }
  }


  return (<><div className="mx-8 mt-10"><div id="main">
    <nav className="bg-white border-gray-200 dark:bg-gray-900">
      <div className=" flex flex-wrap items-center justify-between ">
        <a href="#" className="flex items-center space-x-3 rtl:space-x-reverse">
          <span className="self-center text-2xl font-semibold whitespace-nowrap dark:text-white">CommitCrunch</span>
        </a>
        <div className="flex md:order-2">
        </div>
      </div>
    </nav>

  </div>
    <div>
      <div className="my-4">
        <label htmlFor="default-input" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Enter your GitHub name</label>
        <form onSubmit={e=>onSubmit(e, setData)}>
          <input type="text" id="default-input"
            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
            value={name}
            name="name"
            onChange={(e) => setName(e.target.value)}>
          </input>
          <button type="submit" className="mt-3 w-full text-white bg-blue-500 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm py-2 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">Search</button>
        </form>
      </div>
    </div>
    <div>
      { data && <Insights data={data}/>}
    </div>
  </div></>);
}
