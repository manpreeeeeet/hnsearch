import { useEffect, useState } from "react";
import { search, SearchResult } from "./api.ts";
import Highlighter from "react-highlight-words";

const noResultsArt = `      |\\      _,,,---,,_
ZZZzz /,\`.-'\`'    -.  ;-;;,_
     |,4-  ) )-,_. , (  \`'-'
    '---''(_/--'  \`-'\\_)  No results found`;

const startSearchArt = `      |\\      _,,,---,,_
ZZZzz /,\`.-'\`'    -.  ;-;;,_
     |,4-  ) )-,_. , (  \`'-'
    '---''(_/--'  \`-'\\_)  Search for something`;

function App() {
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [query, setQuery] = useState("");
  const [queryTokens, setQueryTokens] = useState<string[]>([]);
  const [art, setArt] = useState<string>(startSearchArt);

  useEffect(() => {
    setQueryTokens(
      query.split(/(\s+)/).filter(function (e) {
        return e.trim().length > 0;
      }),
    );
    if (searchResults.length === 0 && query.length !== 0) {
      setArt(noResultsArt);
    }
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
      <div className="container mx-auto lg:max-w-screen-lg">
        <div className="bg-[#FF742B] h-12 flex p-2 gap-2 items-center justify-between">
          <div className="font-bold text-sm md:text-xl">HN Search</div>
          <div className="flex whitespace-nowrap">
            <input
              type={"text"}
              className="h-[2rem] w-[200px] md:w-[500px] p-2"
              onInput={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  searchClick();
                }
              }}
            />
            <button
              onClick={searchClick}
              className="border-2 h-[2rem] py-1 px-4 font-bold"
            >
              Search
            </button>
          </div>
        </div>
        <div className="bg-[#F6F6EF] min-h-[calc(100vh_-_6rem_-2px)]">
          <div className="w-full h-full p-2">
            {searchResults.length === 0 && (
              <div className="w-full h-[calc(100vh_-_6rem_-2px)] flex items-center justify-center">
                <textarea
                  value={art}
                  readOnly={true}
                  className="select-none mx-auto"
                  style={{
                    width: "25rem",
                    height: "10rem",
                    border: "none",
                    resize: "none",
                    outline: "none",
                    background: "transparent",
                    fontFamily: "monospace",
                  }}
                ></textarea>
              </div>
            )}
            <div className="p">
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
          <div className="h-[2px] bg-[#FF742B] w-full"></div>
          <div className="w-full h-12 bg-[#F6F6EF] flex items-center justify-center py-4 text-sm gap-2">
            <div className="">
              <span className="text-[#828282]">Made by</span>
              <a href="https://mnprt.me"> @manpreet</a>
            </div>
            <div>
              | <a href="https://github.com/manpreeeeeet/hnsearch">Source</a>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

export default App;
