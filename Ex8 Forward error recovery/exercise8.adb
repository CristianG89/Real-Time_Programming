with Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Float_Random;
use  Ada.Text_IO, Ada.Integer_Text_IO, Ada.Numerics.Float_Random;

procedure exercise8 is
    Count_Failed    : exception;    -- Exception to be raised ("thrown" in C) when counting fails
    Gen             : Generator;    -- Random number generator

	
    -- Declaration of generic functions ----------------------------------------------------------------------
    function Unreliable_Slow_Add (x : Integer) return Integer is
		Error_Rate : Constant := 0.15;  -- (between 0 and 1)
    begin
        if Random(Gen) > Error_Rate then		-- If random number above Error_Rate, I do some operation
            delay Duration(3.5 + Random(Gen));
            return x + 10;
        else
            delay Duration(0.5 * Random(Gen));	-- If random number below, I raise (throw in C/Java) an error
            raise Count_Failed;
        end if;
    end Unreliable_Slow_Add;
	----------------------------------------------------------------------------------------------------------
	
	
	-- Creation (TYPE for declaration and BODY for its implementation) of a thread called TRANSACTION WORKER -
	-- A Task is a program unit that is obeyed concurrently with the rest of an Ada program ("thread" in Java)
    task type Transaction_Worker (Initial : Integer; Manager : access Transaction_Manager);
	----------------------------------------------------------------------------------------------------------
    task body Transaction_Worker is
        Num         : Integer   := Initial;
        Prev        : Integer   := Num;
        Round_Num   : Integer   := 0;
    begin
        Put_Line ("Worker" & Integer'Image(Initial) & " started");

        loop		-- Infinite loop, showing the current round what it is working
            Put_Line ("Worker" & Integer'Image(Initial) & " started round" & Integer'Image(Round_Num));
            Round_Num := Round_Num + 1;

            -------------------------------------------
            -- PART 1: Select-Then-Abort
            select								-- The structure for Asynchronous Transfer of Control (ATC) is the "Select-then-abort" statement,
                Manager.Wait_Until_Aborted;		-- which makes all threads abort and come here, if one of them fails...
                Num := Prev + 5;	-- When process fails, the number is increased in 5 units
                Put_Line ("  Worker" & Integer'Image(Initial) & " committing" & Integer'Image(Num));
            then abort
                begin
                    Num := Unreliable_Slow_Add(Prev);
                    Put_Line ("  Worker" & Integer'Image(Initial) & " committing" & Integer'Image(Num));
                exception
                    when Count_Failed =>
                        Manager.Signal_Abort;
                end;
                Manager.Finished;
            end select;
  
            Prev := Num;	-- Current number becomes the previous one (for the next cycle)
            delay 0.5;		-- Delay for the infinite loop
	        -------------------------------------------	
        end loop;
    end Transaction_Worker;
	----------------------------------------------------------------------------------------------------------

	
	-- Creation (TYPE for declaration and BODY for its implementation) of a thread called TRANSACTION MANAGER -
    protected type Transaction_Manager (N : Positive) is
        entry Finished;
        entry Wait_Until_Aborted;
        procedure Signal_Abort;
    private
        Finished_Gate_Open  : Boolean := False;
        Aborted             : Boolean := False;
    end Transaction_Manager;
	---------------------------------------------------------------------------------------------------------
    protected body Transaction_Manager is
        entry Finished when Finished_Gate_Open or Finished'Count = N is
        begin
            ------------------------------------------
            -- PART 3: Modify the Finished entry
            Finished_Gate_Open := Finished'Count /= 0;		-- /= compares (like ==)
            if not Finished_Gate_Open then
                Aborted := False;
            end if;
            ------------------------------------------
        end Finished;

        -------------------------------------------
        -- PART 2: Wait_Until_Aborted
        entry Wait_Until_Aborted when Aborted is
        begin 
            if Wait_Until_Aborted'Count = 0 then
                Aborted := False;
            end if;
        end; 
        -------------------------------------------

        procedure Signal_Abort is
        begin
            Aborted := True;
        end Signal_Abort;

    end Transaction_Manager;
	----------------------------------------------------------------------------------------------------------	
	
	
	-- Instantiation of objects for executing the assigned tasks/threads (including the defined arguments) ---
    Manager : aliased Transaction_Manager (3);

    Worker_1 : Transaction_Worker (0, Manager'Access);
    Worker_2 : Transaction_Worker (1, Manager'Access);
    Worker_3 : Transaction_Worker (2, Manager'Access);
	----------------------------------------------------------------------------------------------------------
	
begin
    Reset(Gen);		-- Seed the random number generator
end exercise8;