import { useEffect, useState } from "react";
import { search, SearchResult } from "./api.ts";

function App() {
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [query, setQuery] = useState("");

  useEffect(() => {
    search(query).then((data) => {
      if (data) {
        console.log("search results: ", data);
        setSearchResults(data);
      }
    });
  }, [query]);

  return (
    <>
      <div className="container mx-auto lg:max-w-screen-lg h-screen">
        <div className="bg-[#FF742B] h-16 flex p-2 gap-2 items-center">
          <div>Search</div>
          <input
            type={"text"}
            className="h-[2rem] p-2"
            onInput={(event) => setQuery(event.target.value)}
          />
        </div>
        <div className="bg-[#F6F6EF] h-full">
          <div>
            {searchResults.map((searchResult) => {
              return (
                <div key={searchResult.id}>{searchResult.story.title}</div>
              );
            })}
          </div>
        </div>
      </div>
    </>
  );
}

export default App;
