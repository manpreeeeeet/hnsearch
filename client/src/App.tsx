import React, { useEffect, useState } from "react";
import { search, SearchResult } from "./api.ts";
import Highlighter from "react-highlight-words";

function App() {
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [query, setQuery] = useState("");
  const [queryTokens, setQueryTokens] = useState<string[]>([]);

  useEffect(() => {
    setQueryTokens(
      query.split(/(\s+)/).filter(function (e) {
        return e.trim().length > 0;
      }),
    );
  }, [query]);

  const searchClick = () => {
    search(query).then((data) => {
      if (data) {
        setSearchResults(data);
      }
    });
  };

  const regexSearchWords = queryTokens.map(
    (word) => new RegExp(`\\b${word}\\b`, "gi"),
  );

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
          <button onClick={searchClick}>search</button>
        </div>
        <div className="bg-[#F6F6EF] min-h-full p-2">
          <div>
            {searchResults.map((searchResult) => {
              return (
                <div>
                  <div className="text-base my-4" key={searchResult.id}>
                    <Highlighter
                      searchWords={regexSearchWords}
                      textToHighlight={searchResult.story.title}
                    />
                  </div>
                  {searchResult.comments.map((comment) => {
                    return (
                      <>
                        <div className="text-sm border border-slate-700 mt-2 p-2">
                          <Highlighter
                            searchWords={regexSearchWords}
                            textToHighlight={comment.text}
                          />
                        </div>
                      </>
                    );
                  })}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </>
  );
}

export default App;
