import { useEffect, useState } from "react";
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
  }, [searchResults]);

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
          <input
            type={"text"}
            className="h-[2rem] w-[60%] p-2"
            onInput={(event) => setQuery(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                searchClick();
              }
            }}
          />
          <button onClick={searchClick} className="border-2 h-[2rem] py-1 px-4">
            Search
          </button>
        </div>
        <div className="bg-[#F6F6EF] min-h-full p-2">
          <div>
            {searchResults.map((searchResult) => {
              return (
                <div>
                  <div className="text-base my-4" key={searchResult.id}>
                    <div>
                      <a href={searchResult.story.url}>
                        <Highlighter
                          searchWords={regexSearchWords}
                          textToHighlight={searchResult.story.title}
                        />
                      </a>
                      <div className="text-xs text-gray-600">
                        {searchResult.story.score} points
                      </div>
                    </div>
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
