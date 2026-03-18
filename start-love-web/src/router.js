import {createRouter, createWebHistory} from "vue-router";
import {httpPost} from "@/utils/http";
import {showMessageError} from "@/utils/dialog";

const routes = [
    {
        name: "Index",
        path: "/",
        redirect: "/home",
        meta: {title: "首页"},
        component: () => import("@/views/Index.vue"),
    },
    {
        name: "home",
        path: "/home",
        redirect: "/chat",
        component: () => import("@/views/Home.vue"),
        children: [
            {
                name: "chat",
                path: "/chat",
                meta: {title: "星启诶艾"},
                component: () => import("@/views/Chat.vue"),
            },
            {
                name: "chat-id",
                path: "/chat/:id",
                meta: {title: "星启诶艾"},
                component: () => import("@/views/Chat.vue"),
            },
            {
                name: "report",
                path: "/report",
                meta: {title: "我的报告"},
                component: () => import("@/views/Report.vue"),
            },
            {
                name: "single-technology",
                path: "/single-page/early-screening-technology",
                meta: {title: "早筛技术"},
                component: () => import("@/views/SingleTechnology.vue"),
            },
            {
                name: "single-policy",
                path: "/single-page/policy-guide",
                meta: {title: "政策指引"},
                component: () => import("@/views/SinglePolicy.vue"),
            },
            {
                name: "single-star",
                path: "/single-page/star-love",
                meta: {title: "星启协爱"},
                component: () => import("@/views/SingleFamily.vue"),
            },
            {
                name: "single-institution",
                path: "/single-page/public-institution",
                meta: {title: "公益机构"},
                component: () => import("@/views/SingleInstitution.vue"),
            },
            {
                name: "single-platform",
                path: "/single-page/open-platform",
                meta: {title: "开放平台"},
                component: () => import("@/views/SinglePlatform.vue"),
            },
            {
                name: "single-about",
                path: "/single-page/about",
                meta: {title: "关于我们"},
                component: () => import("@/views/SingleAbout.vue"),
            },
        ]
    },
    {
        name: "login",
        path: "/login",
        meta: {title: "用户登录"},
        component: () => import("@/views/Login.vue"),
    },
    {
        name: "wechat-login-callback",
        path: "/login/wechat/callback",
        meta: {title: "用户登录"},
        component: () => import("@/views/LoginCallback.vue"),
    },
    {
        name: "mobile-login",
        path: "/mobile/login",
        meta: {title: "用户登录"},
        component: () => import("@/views/mobile/Login.vue"),
    },
    {
        path: "/mobile/chat/:id",
        name: "mobile-chat-id",
        meta: {title: "我的会话"},
        component: () => import("@/views/mobile/Chat.vue"),
    },
    {
        name: "mobile",
        path: "/mobile",
        meta: {title: "首页"},
        component: () => import("@/views/mobile/Home.vue"),
        redirect: "/mobile/chat/add",
        children: [
            {
                path: "/mobile/chat/add",
                name: "mobile-chat",
                meta: {title: "新建会话"},
                component: () => import("@/views/mobile/AddChat.vue"),
            },

            {
                path: "/mobile/chat/list",
                name: "mobile-chat-list",
                meta: {title: "绘画列表"},
                component: () => import("@/views/mobile/ChatList.vue"),
            },
            {
                path: "/mobile/report",
                name: "mobile-report",
                meta: {title: "我的报告"},
                component: () => import("@/views/mobile/Report.vue"),
            },
            {
                path: "/mobile/profile",
                name: "mobile-profile",
                meta: {title: "更多"},
                component: () => import("@/views/mobile/Profile.vue"),
            },
            {
                name: "mobile-single-technology",
                path: "/mobile/single-page/early-screening-technology",
                meta: {title: "早筛技术"},
                component: () => import("@/views/mobile/SingleTechnology.vue"),
            },
            {
                name: "mobile-single-policy",
                path: "/mobile/single-page/policy-guide",
                meta: {title: "政策指引"},
                component: () => import("@/views/mobile/SinglePolicy.vue"),
            },
            {
                name: "mobile-single-star",
                path: "/mobile/single-page/star-love",
                meta: {title: "星启协爱"},
                component: () => import("@/views/mobile/SingleFamily.vue"),
            },
            {
                name: "mobile-single-institution",
                path: "/mobile/single-page/public-institution",
                meta: {title: "公益机构"},
                component: () => import("@/views/mobile/SingleInstitution.vue"),
            },
            {
                name: "mobile-single-platform",
                path: "/mobile/single-page/open-platform",
                meta: {title: "开放平台"},
                component: () => import("@/views/mobile/SinglePlatform.vue"),
            },
            {
                name: "mobile-single-about",
                path: "/mobile/single-page/about",
                meta: {title: "关于我们"},
                component: () => import("@/views/mobile/SingleAbout.vue"),
            },
        ],
    },
    {
        name: "NotFound",
        path: "/:all(.*)",
        meta: {title: "页面没有找到"},
        component: () => import("@/views/404.vue"),
    },
];

// console.log(MY_VARIABLE)
const router = createRouter({
    history: createWebHistory(),
    routes: routes,
});

let prevRoute = null;
// dynamic change the title when router change
router.beforeEach((to, from, next) => {
    document.title = to.meta.title;
    prevRoute = from;
    let code = null
    if (to.path === '/wechatLogin') {
        code = to.query.code || null
    }
    // 微信授权登陆
    if (code) {
        httpPost("/api/login/wechat")
            .then((res) => {
                console.log(res);
                router.push({path: '/chat'})
            })
            .catch((e) => {
                showMessageError("登录失败，" + e.message);
            });
    } else {
        next()
    }
});

export {router, prevRoute};
