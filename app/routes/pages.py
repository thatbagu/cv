from fastapi import APIRouter, Request, HTTPException
from fastapi.responses import HTMLResponse, PlainTextResponse
from bs4 import BeautifulSoup
import asyncio
from app.utils import read_file, generate_blog_html
from app.routes.blog import get_blog_posts

router = APIRouter()


@router.get("/pgp", response_class=PlainTextResponse)
async def serve_pgp_key():
    key = await asyncio.to_thread(read_file, "assets/pgp.asc")
    return PlainTextResponse(content=key.decode("utf-8"), media_type="application/pgp-keys")


@router.get("/", response_class=HTMLResponse)
@router.get("/home", response_class=HTMLResponse)
@router.get("/about", response_class=HTMLResponse)
@router.get("/projects", response_class=HTMLResponse)
@router.get("/contact", response_class=HTMLResponse)
@router.get("/blog", response_class=HTMLResponse)
async def serve_page(request: Request):
    path = request.url.path
    try:
        is_htmx_request = request.headers.get("HX-Request")
        content_file = "pages/home.html" if path == "/" else f"pages{path}.html"

        content = await asyncio.to_thread(read_file, content_file)
        content = content.decode("utf-8")

        if path == "/blog":
            posts = await get_blog_posts()
            blog_content = generate_blog_html(posts)
            soup = BeautifulSoup(content, "html.parser")
            blog_posts_div = soup.find(id="blog-posts")
            if blog_posts_div:
                blog_posts_div.clear()
                blog_posts_div.append(BeautifulSoup(blog_content, "html.parser"))
            content = str(soup)

        if is_htmx_request:
            return HTMLResponse(content=content)
        else:
            index_html = await asyncio.to_thread(read_file, "index.html")
            index_html = index_html.decode("utf-8")

            soup = BeautifulSoup(index_html, "html.parser")
            main_content = soup.find(id="main-content")
            if main_content:
                main_content.clear()
                main_content.append(BeautifulSoup(content, "html.parser"))

            return HTMLResponse(content=str(soup))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Internal Server Error: {str(e)}")
