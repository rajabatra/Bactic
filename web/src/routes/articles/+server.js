import json from "@sveltejs/kit";

async function getArticles() {
  var articles;

  const paths = import.meta.glob("/src/articles/*.md", { eager: true });

  for (const path in paths) {
    const file = paths[path];
    const slug = path.split("/").at(-1)?.replace(".md", "");

    if (file && typeof file === "object" && "metadata" in file && slug) {
      const metadata = file.metadata;
      const article = { ...metadata, slug };
      article.published && articles.push(article);
    }
  }

  return articles;
}

export async function GET() {
  let articles = await getArticles();
  return json(articles);
}
