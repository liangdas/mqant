require([
    "gitbook",
    "jQuery"
], function(gitbook, $) {
    var mygitalk = {};

    gitbook.events.bind("start", function(e, config) {
        mygitalk = config.mygitalk || {};

        initMygitalk();
    });

    gitbook.events.bind("page.change", initMygitalk);

    function initMygitalk() {
        mygitalk.id = md5(location.pathname);

        var gitalk = new Gitalk(mygitalk);
        gitalk.render("gitalk-container");
    }
});