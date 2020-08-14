/*
Copyright © 2020 gikoha

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-oci8"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type ColumnAttr struct {
	Name    string
	Attr    string // eg. VARCHAR(20)
	Default string
	NotNull bool
}

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "SQLbackup <user/password@XEPDB1>",
	Short: "Backup oracle DB into SQL file",
	Long:  `Backup all oracle database that belongs to user into SQL file, including BLOB.`,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires credentials")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		analyze(args[0])
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var droptable bool

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.SQLbackup.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolVarP(&droptable, "drop", "d", false, "DROP table beforce CREATE")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".SQLbackup" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".SQLbackup")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// ddlAnalyze  analyze DDL string
func ddlAnalyze(db *sql.DB, tablename string, ddl string) {

	sc := bufio.NewScanner(strings.NewReader(ddl))

	for i := 1; sc.Scan(); i++ {
		if err := sc.Err(); err != nil {
			panic(err)
			// エラー処理
		}
		str := sc.Text()
		s := strings.TrimSpace(str)

		if strings.HasPrefix(s, "CREATE TABLE") {
			// CREATE TABLE
			var tableColumns []ColumnAttr

			for ; sc.Scan(); i++ {
				s = strings.TrimSpace(sc.Text())
				if strings.HasPrefix(s, "(") {
					s = strings.TrimSpace(s[2:])
				}

				if strings.HasPrefix(s, ")") {
					break
				}
				if !strings.HasPrefix(s, "\"") {
					// not started with double quote .. USING...
					continue
				}
				s = strings.TrimRight(s, ",")
				slice := strings.SplitN(s, " ", 2)
				var oneColumn ColumnAttr
				oneColumn.Name = slice[0]
				s = slice[1]
				oneColumn.NotNull = false
				if strings.Contains(s, "NOT NULL ENABLE") {
					s = strings.Replace(slice[1], "NOT NULL ENABLE", "", 1)
					oneColumn.NotNull = true
				}
				oneColumn.Default = ""
				if strings.Contains(s, "DEFAULT") {
					s2 := strings.SplitN(s, "DEFAULT", 2)
					ss := strings.TrimSpace(s2[1])
					if strings.HasPrefix(ss, "'") {
						s3 := strings.SplitN(ss, "'", 3)
						ss = "'" + s3[1] + "'"
						s3[1] = ""
						s = s2[0] + strings.Join(s3, " ")
					} else {
						s3 := strings.SplitN(ss, " ", 2)
						ss = s3[0]
						s3[0] = ""
						s = s2[0] + strings.Join(s3, " ")
					}
					oneColumn.Default = ss

				}

				oneColumn.Attr = s

				tableColumns = append(tableColumns, oneColumn)
			}

			// make Insert
			s1 := ""
			for _, s := range tableColumns {
				if strings.HasPrefix(s.Attr, "NVARCHAR") {
					s1 += s.Name + ","
				} else if strings.HasPrefix(s.Attr, "BLOB") {
					s1 += s.Name + ","
				} else { // number or date
					s1 += "TO_CHAR(" + s.Name + "),"
				}
			}
			s1 = strings.TrimRight(s1, ",")
			querystr := "SELECT " + s1 + " FROM " + tablename + "\n"
			//fmt.Println(querystr)

			ctx := context.Background()

			rows, err := db.QueryContext(ctx, querystr)
			if err != nil {
				panic(err)
			}
			cols, err := rows.Columns()
			if err != nil {
				panic(err)
			}

			// 参考 https://qiita.com/taizo/items/54f5f49c6102f86194b8
			// 任意の型、数のColumnsを読み込む

			fmt.Println("SET DEFINE OFF;")
			values := make([]sql.RawBytes, len(cols))
			scaninterface := make([]interface{}, len(values))
			for i := range values {
				scaninterface[i] = &values[i]
			}
			for rows.Next() {
				s1 = ""
				err = rows.Scan(scaninterface...)
				if err != nil {
					panic(err)
				}
				for i, col := range values {
					// Here we can check if the value is nil (NULL value)
					if col == nil {
						s1 += "NULL,"
					} else if strings.HasPrefix(tableColumns[i].Attr, "NVARCHAR") {
						s1 += "'" + string(col) + "',"
					} else if strings.HasPrefix(tableColumns[i].Attr, "DATE") {
						s1 += "TO_DATE('" + string(col) + "','RR-MM-DD'),"
					} else if strings.HasPrefix(tableColumns[i].Attr, "BLOB") {
						s1 += "HEXTORAW('" + hex.EncodeToString(col) + "'),"
					} else {
						s1 += string(col) + ","
					}
				}
				s2 := "Insert Into " + tablename + " ("
				for _, s := range tableColumns {
					s2 += s.Name + ","
				}
				s1 = strings.TrimRight(s1, ",")
				s2 = strings.TrimRight(s2, ",")

				s2 += ") VALUES (" + s1 + ");"
				fmt.Println(s2)
			}
			rows.Close()

			continue
		}

	}

	fmt.Println("")
	fmt.Println("")

}

// tableAnalyze : analyze table
func tableAnalyze(db *sql.DB, tablename string) {

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)

	err := db.PingContext(ctx)
	cancel()
	if err != nil {
		panic(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var rows *sql.Rows
	rows, err = db.QueryContext(ctx, "select dbms_metadata.get_ddl('TABLE',:1) from dual", tablename)

	if err != nil {
		panic(err)
	}
	var ddl string = ""
	for rows.Next() {
		var row string
		err = rows.Scan(&row)
		if err != nil {
			panic(err)
		}
		ddl += row + ";\n"
	}
	rows.Close()

	if droptable {
		fmt.Printf("DROP TABLE \"%s\";\n", tablename)
	}
	fmt.Println(ddl)
	ddlAnalyze(db, tablename, ddl)
}

// analyze : analyze user
func analyze(credential string) {
	db, err := sql.Open("oci8", credential)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	ctx := context.Background()
	var rows *sql.Rows

	rows, err = db.QueryContext(ctx, "select table_name from user_tables")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var tablename string
		err = rows.Scan(&tablename)
		if err != nil {
			panic(err)
		}
		tableAnalyze(db, tablename)
	}
	rows.Close()

}
