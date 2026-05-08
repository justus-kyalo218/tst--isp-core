db = db.getSiblingDB("tst_isp");

db.createCollection("users");
db.createCollection("sub_isps");
db.createCollection("payments");
db.createCollection("routers");

print("Initialized tst_isp Mongo database");
