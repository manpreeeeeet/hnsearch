export interface SearchResult {
  id: number;
  story: Story;
  comments: Comment[];
}

export interface Story {
  title: string;
  url: string;
  score: number;
}

export interface Comment {
  text: string;
}

export const search = async (query: string) => {
  try {
    const response = await fetch(`/search?q=${query}`);
    const data: SearchResult[] = await response.json();
    return data;
  } catch (error) {
    console.error("Error:", error);
  }
};
