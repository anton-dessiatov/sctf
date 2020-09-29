-- +goose Up
CREATE TABLE `elastic_ip` (
  id INT NOT NULL AUTO_INCREMENT,
  cluster_id INT NOT NULL,
  external_id VARCHAR(50) NOT NULL,
  instance_id INT NULL,
  association_id VARcHAR(50) NULL,
  status VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT NOW(),
  updated_at DATETIME NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
