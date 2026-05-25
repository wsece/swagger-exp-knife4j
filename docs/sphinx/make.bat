@ECHO OFF
setlocal

set SPHINXBUILD=sphinx-build
set SOURCEDIR=.
set BUILDDIR=_build

if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="sync" goto sync
if "%1"=="html" goto html
if "%1"=="clean" goto clean
if "%1"=="serve" goto serve

%SPHINXBUILD% -M %1 %SOURCEDIR% %BUILDDIR%
goto end

:help
%SPHINXBUILD% -M help %SOURCEDIR% %BUILDDIR%
goto end

:sync
python sync_md.py
goto end

:html
python sync_md.py
%SPHINXBUILD% -M html %SOURCEDIR% %BUILDDIR%
python postprocess_sourcelinks.py
goto end

:clean
if exist %BUILDDIR% rmdir /s /q %BUILDDIR%
if exist source_md rmdir /s /q source_md
goto end

:serve
call :html
echo Open http://127.0.0.1:8000/
python serve_docs.py --port 8000 --directory %BUILDDIR%\html
goto end

:end
endlocal
