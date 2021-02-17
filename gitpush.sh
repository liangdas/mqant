cd `pwd`/_book
git init
git checkout --orphan gh-pages
git add .
git commit -m "init project"
git remote add origin https://github.com/liangdas/mqant
git push origin gh-pages