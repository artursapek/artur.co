require 'date'
require 'optparse'

album_file = ARGV[0]
caption_style = "before"


OptionParser.new do |opts|
  opts.on("-c", "--captions [LOCATION]", String, "Caption style") do |v|
    puts v
    if v != "before" && v != "after"
      puts "fatal: --captions option must be either \"before\" or \"after\""
      exit 1
    end
    caption_style = v
  end
end.parse!

title = nil
date = nil
images = []

pending_caption = nil

puts "Parsing album at #{album_file}"

File.open(album_file, "r") do |f|
  f.each_line do |line|
    if line =~ /h1/
      title = line.gsub(/<\/?h1>/,"").split.map(&:capitalize).join(' ')
    elsif line =~ /h3/
      date = Date.parse(line.gsub(/<\/?h3>/,"").split.map(&:capitalize).join(' '))
    elsif line =~ /<p[a-z=\-\"\ ]*>/
      caption = line.gsub(/<\/?p[a-z=\-\"\ ]*>/,"").gsub("\n","").gsub("\"", "\\\"")

      if caption_style == "before"
        pending_caption = caption
      else
        images[-1][:caption] = caption
      end

    elsif line =~ /video.*src/
      src = /src="([^"]*)"/.match(line)[1]
      if src =~ /http/
        fn = src
      else
        src.gsub!("./","")
        fn = src.split(".")[0]
      end
      video = { type: "video", fn: fn }
      if pending_caption
        video[:caption] = pending_caption
        pending_caption = nil
      end
      images << video

    elsif line =~ /img.*src/
      src = /src="([^"]*)"/.match(line)[1]
      src.gsub!("./","")
      fn = src.split(".")[0]
      image = { type: "photo", fn: fn }
      if pending_caption
        image[:caption] = pending_caption
        pending_caption = nil
      end
      images << image

    end
  end
end

yml_file = [
  "title: #{title}",
  "date: #{date.strftime "%-m/%-d/%y"}",
  "content:"
]

images.each do |img|
  fn = img[:fn]

  if img[:type] == "photo"

    paths = `sshb artur "cd /mnt/raw/photos; find . -name *#{fn}*"`.split("\n")
    # For the purposes of this script, we can omit anything
    # before 2014 to eliminate duplicate filenames
    paths = paths.select{ |p| p.split("/")[1].to_i > 2013 }

    if paths.size > 1
      puts "Ambiguous fn #{fn}; more than one found path"
      puts paths.map{|p| "- #{p}"}
      puts ""
    elsif paths.size == 0
      puts "Cannot find raw path for #{fn}"
    else
      path = paths[0]
      if path == ""
        puts "Cannot find raw path for #{fn}"
      else
        year = path.split("/")[1]
        month = path.split("/")[2]
        yml_file << "- type: #{img[:type]}"
        yml_file << "  src: #{[year, month, "#{fn}.JPG"].join("/")}"
        if img[:caption]
          yml_file << "  caption: \"#{img[:caption]}\""
        end
      end
    end
  else
    path_to_video = nil
    bn = File.basename fn
    if fn =~ /http/
      # wget it
      `sshb artur "cd /tmp/; wget #{fn}"`
      path_to_video = "/tmp/#{bn}"
    else
      # cp it
      results = `sshb artur "cd /app/artur/photos; find . -name *#{fn}.mov*"`.split("\n")
      bn = "#{bn}.mov"
      if results.size == 0
        puts "Couldn't find video #{fn}"
        next
      else
        path_to_video = results[0].strip
      end
    end

    puts "FN #{fn}"
    puts "RP #{path_to_video}"

    date = Date.parse(
      `sshb artur "cd /app/artur/photos; ffmpeg -i #{path_to_video} 2>&1 | grep creation_time | tail -n1 | cut -d ':' -f 2"`
      .strip.split(" ")[0]
    )
    puts "DATE #{date}"

    dir = "/mnt/raw/videos/#{date.year}/#{date.month.to_s.rjust(2, '0')}/"
    puts "DIR #{dir}"

    `sshb artur "cd /app/artur/photos; mkdir -p #{dir}; cp #{path_to_video} #{dir}"`

    `sshb artur "rm #{path_to_video}"` if fn =~ /http/

    yml_file << "- type: #{img[:type]}"
    yml_file << "  src: #{date.year}/#{date.month.to_s.rjust(2, '0')}/#{bn}"
    if img[:caption]
      yml_file << "  caption: \"#{img[:caption]}\""
    end
  end
end

if title.nil?
  raise "Couldn't find title h1 element!"
elsif date.nil?
  raise "Couldn't find date h3 element!"
end

slug = title.split(" ").map(&:downcase).join("-")

f = File.new("/usr/local/go/src/github.com/artursapek/artur.co/content/photos/albums/#{slug}.yml", "w")
f.write(yml_file.join("\n"))
f.close
