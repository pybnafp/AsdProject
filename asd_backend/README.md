# 初始化

go install github.com/beego/bee/v2@latest

## 依赖库初始化

go mod tidy

## 数据库初始化

```sql mysql
CREATE DATABASE IF NOT EXISTS `asd` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE USER 'asd'@'%' IDENTIFIED BY 'Pt7Dz1OIaR0TJsKb';
GRANT ALL PRIVILEGES ON `asd`.* TO 'asd'@'%';
FLUSH PRIVILEGES;
```

```sql postgres
create database asd;
create user asd with encrypted password 'Pt7Dz1OIaR0TJsKb';
grant all privileges on database asd to asd;
```
